package connect_deployment

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/connect"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

const (
	Sandbox    = "sandbox"
	Production = "production"
	Preview    = "preview"
)

// pollInterval is how often a deployment's status is read while waiting for
// Connect to build, deploy or undeploy it.
var pollInterval = 15 * time.Second

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *connect.ByProjectKeyRequestBuilder
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connect_deployment"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	configuration := func(description string, sensitive bool) schema.MapAttribute {
		return schema.MapAttribute{
			MarkdownDescription: description,
			ElementType:         types.StringType,
			Optional:            true,
			Sensitive:           sensitive,
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A deployment of a commercetools Connect connector to the Project: its " +
			"applications built and run by commercetools, with the configuration given here.\n\n" +
			"Terraform uploads no code. The connector names a Git repository and tag; Connect builds the " +
			"applications from that tag. This resource installs the connector in the Project with its " +
			"configuration; roll out new releases from CI/CD with the commercetools CLI, which publishes " +
			"a new tag and redeploys with `--updateConnector`, and add `connector_version` to " +
			"`ignore_changes`.\n\n" +
			"Creating one waits until Connect has deployed it, which usually takes several minutes. A " +
			"deployment that fails stays in the state as tainted, so the next apply replaces it. A change to " +
			"the configuration or the connector version redeploys it in place; a change to the connector, the " +
			"type or the region replaces it.\n\n" +
			"Secured configuration values are never returned. Terraform keeps the configured values in state, " +
			"cannot detect that they were changed outside of Terraform, and sends them again on every redeploy. " +
			"After an import, set them before the first apply, or the redeploy sends none.\n\n" +
			"See also the [Connect deployments documentation](https://docs.commercetools.com/connect/deployments).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the deployment.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description:   "User-defined unique identifier of the deployment.",
				Required:      true,
				PlanModifiers: replace,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the deployment, raised by every redeploy.",
				Computed:    true,
			},
			"connector_key": schema.StringAttribute{
				MarkdownDescription: "Key of the connector to deploy. Conflicts with `connector_id`.",
				Optional:            true,
				PlanModifiers:       replace,
				Validators: []validator.String{
					stringvalidator.ExactlyOneOf(path.MatchRoot("connector_key"), path.MatchRoot("connector_id")),
				},
			},
			"connector_id": schema.StringAttribute{
				MarkdownDescription: "ID of the connector to deploy. Conflicts with `connector_key`.",
				Optional:            true,
				PlanModifiers:       replace,
			},
			"connector_version": schema.Int64Attribute{
				MarkdownDescription: "Version of the connector to deploy. Defaults to its current version, " +
					"and then holds the version that runs. When releases are rolled out with the commercetools " +
					"CLI (`commercetools connect deployment redeploy --updateConnector`), the recommended way, " +
					"add it to `ignore_changes`. Raising it redeploys in place, moving the deployment to the " +
					"connector's latest published version, which must be this one.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"connector_staged": schema.BoolAttribute{
				MarkdownDescription: "Deploy the connector's previewable staged version instead of the " +
					"published one. Requires `type = \"preview\"`.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "`sandbox` or `production` for a published connector, `preview` for a " +
					"staged one. Defaults to `sandbox`.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf(Sandbox, Production, Preview),
				},
			},
			"region": schema.StringAttribute{
				MarkdownDescription: "Region to deploy to, such as `europe-west1.gcp`.",
				Required:            true,
				PlanModifiers:       replace,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"global_standard_configuration": configuration(
				"Plain values for the connector's global configuration, shared by every application.", false),
			"global_secured_configuration": configuration(
				"Secret values for the connector's global configuration, shared by every application.", true),
			"skip_scripts_on_redeploy": schema.BoolAttribute{
				MarkdownDescription: "Skip the connector's `postDeploy` and `preUndeploy` scripts when a change " +
					"redeploys it.",
				Optional: true,
			},
			"deploy_timeout": schema.StringAttribute{
				MarkdownDescription: "How long to wait for a deployment, redeployment or undeployment, as a " +
					"duration such as `45m`. Defaults to `30m`.",
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("30m"),
				Validators: []validator.String{
					durationValidator{},
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the deployment, `Deployed` once it runs.",
				Computed:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"application": schema.ListNestedBlock{
				MarkdownDescription: "Configuration of one of the connector's applications, named as in its " +
					"`connect.yaml`. Deploy at least one.",
				Validators: []validator.List{
					listvalidator.IsRequired(),
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the application, as in the connector's `connect.yaml`.",
							Required:            true,
						},
						"standard_configuration": configuration(
							"Plain values, read by the application as environment variables.", false),
						"secured_configuration": configuration(
							"Secret values, read by the application as environment variables.", true),
						"schedule": schema.StringAttribute{
							MarkdownDescription: "Cron schedule of a `job` application, overriding the one in " +
								"`connect.yaml`.",
							Optional: true,
						},
						"url": schema.StringAttribute{
							Description: "URL of a service application, once deployed.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"topic": schema.StringAttribute{
							Description: "Google Cloud Pub/Sub topic of an event application, once deployed.",
							Computed:    true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (r *Resource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, ok := req.ProviderData.(*utils.ProviderData)
	if !ok {
		return
	}
	r.client = data.ConnectClient
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Deployment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft, diags := plan.draft(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Deployments().Post(draft).Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error creating connect deployment", processError(err).Error())
		return
	}

	// From here on the deployment exists, so it goes into the state whatever
	// the outcome of the wait; a failed one is then tainted and replaced.
	res, waitErr := r.waitUntilDeployed(ctx, res, plan.DeployTimeout)
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
	if waitErr != nil {
		resp.Diagnostics.AddError("Error creating connect deployment", waitErr.Error())
	}
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Deployment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Deployments().WithKey(state.Key.ValueString()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading connect deployment", processError(err).Error())
		return
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, state)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Deployment
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	action, diags := plan.redeploy(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Deployments().WithKey(state.Key.ValueString()).Post(connect.DeploymentUpdate{
		Version: int(state.Version.ValueInt64()),
		Actions: []connect.DeploymentUpdateAction{action},
	}).Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error updating connect deployment", processError(err).Error())
		return
	}

	res, waitErr := r.waitUntilDeployed(ctx, res, plan.DeployTimeout)
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
	if waitErr != nil {
		resp.Diagnostics.AddError("Error updating connect deployment", waitErr.Error())
		return
	}
	if want := plan.ConnectorVersion; !want.IsUnknown() && !want.IsNull() && int64(res.Connector.Version) != want.ValueInt64() {
		resp.Diagnostics.AddAttributeError(path.Root("connector_version"), "Error updating connect deployment",
			fmt.Sprintf("Connect moved the deployment to version %d, the connector's latest published version, "+
				"not to version %d: a redeploy can only move it to the latest one.", res.Connector.Version, want.ValueInt64()))
	}
}

// Delete undeploys the deployment and waits until Connect has removed it.
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Deployment
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := state.Key.ValueString()
	_, err := r.client.Deployments().WithKey(key).Delete().Execute(ctx)
	if err != nil && !utils.IsResourceNotFoundError(err) {
		resp.Diagnostics.AddError("Error deleting connect deployment", processError(err).Error())
		return
	}

	err = r.poll(ctx, state.DeployTimeout, func() (bool, error) {
		res, err := r.client.Deployments().WithKey(key).Get().Execute(ctx)
		if utils.IsResourceNotFoundError(err) {
			return true, nil
		}
		if err != nil {
			return false, processError(err)
		}
		if res.Status == string(connect.DeploymentStatusUndeployFailed) {
			return false, fmt.Errorf("connect could not undeploy %q", key)
		}
		return false, nil
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting connect deployment", err.Error())
	}
}

// ImportState imports a deployment by its key. Secured configuration values
// are not returned, so they are empty after an import.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("deploy_timeout"), "30m")...)
}

func (r *Resource) setState(ctx context.Context, state stateSetter, diags *diag.Diagnostics, res *connect.Deployment, prior Deployment) {
	current, d := NewDeploymentFromNative(ctx, res, prior)
	diags.Append(d...)
	diags.Append(state.Set(ctx, current)...)
}

type stateSetter interface {
	Set(ctx context.Context, val any) diag.Diagnostics
}

// waitUntilDeployed returns the deployment once Connect has deployed it, or
// an error when the deployment failed or did not finish in time. Either way
// it returns the last deployment read.
func (r *Resource) waitUntilDeployed(ctx context.Context, res *connect.Deployment, timeout types.String) (*connect.Deployment, error) {
	last := res
	err := r.poll(ctx, timeout, func() (bool, error) {
		// Every other status is on its way to one of these two.
		status := connect.DeploymentStatus(last.Status)
		if status == connect.DeploymentStatusDeployed {
			return true, nil
		}
		if status == connect.DeploymentStatusFailed {
			return false, fmt.Errorf("connect could not deploy %q: the deployment failed, see "+
				"`commercetools connect deployment logs` for why", last.ID)
		}
		next, err := r.client.Deployments().WithId(last.ID).Get().Execute(ctx)
		if err != nil {
			return false, processError(err)
		}
		last = next
		return false, nil
	})
	return last, err
}

var errTimeout = errors.New("timed out")

// poll calls done until it reports true, returns an error, or the timeout
// passes.
func (r *Resource) poll(ctx context.Context, timeout types.String, done func() (bool, error)) error {
	limit, err := time.ParseDuration(timeout.ValueString())
	if err != nil {
		limit = 30 * time.Minute
	}
	deadline := time.Now().Add(limit)
	for {
		finished, err := done()
		if err != nil || finished {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%w after %s; raise deploy_timeout to wait longer", errTimeout, limit)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

func processError(err error) error {
	if retryErr := utils.ProcessRemoteError(err); retryErr != nil && retryErr.Err != nil {
		return retryErr.Err
	}
	return err
}

type durationValidator struct{}

func (durationValidator) Description(_ context.Context) string {
	return "must be a duration such as 30m"
}

func (v durationValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v durationValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if d, err := time.ParseDuration(req.ConfigValue.ValueString()); err != nil || d <= 0 {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid duration", v.Description(ctx))
	}
}
