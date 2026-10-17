package connect_connector

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
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

// projectIdentifier is a Project as Connect names it, {region}:{projectKey}.
var projectIdentifier = regexp.MustCompile(`^[a-zA-Z0-9-_.]+:[a-zA-Z0-9-_]{2,36}$`)

// pollInterval is how often the connector is read while waiting for Connect
// to publish it.
var pollInterval = 15 * time.Second

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *connect.ConnectorsRequestBuilder
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connect_connector"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	optionalText := func(description string) schema.StringAttribute {
		return schema.StringAttribute{
			MarkdownDescription: description,
			Optional:            true,
			Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
		}
	}
	stringSet := func(description string, validators ...validator.String) schema.SetAttribute {
		return schema.SetAttribute{
			MarkdownDescription: description,
			ElementType:         types.StringType,
			Optional:            true,
			Validators:          []validator.Set{setvalidator.ValueStringsAre(validators...)},
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A commercetools Connect connector: the Git repository and tag that Connect builds " +
			"its applications from, and who may deploy it. This is the connector's staged version; with " +
			"`publish` set, Terraform also publishes it, privately and without certification, to the " +
			"projects in `private_projects`. Install it in a Project with a `commercetools_connect_deployment`.\n\n" +
			"A private repository must be given by its SSH URL, and the `connect-mu` GitHub user needs read " +
			"access to it.\n\n" +
			"Releases are usually made outside of Terraform, by setting a new tag and publishing with the " +
			"commercetools CLI (`commercetools connect connectorstaged update --repository-tag` and " +
			"`commercetools connect connectorstaged publish`). Add `repository_tag` to `ignore_changes` so " +
			"Terraform keeps the tag the CLI set, the way it would keep a container image a deployment " +
			"pipeline set. The same CI/CD step then redeploys the deployments with " +
			"`commercetools connect deployment redeploy --updateConnector`.\n\n" +
			"See also the [Connectors documentation](https://docs.commercetools.com/connect/connectors-staged).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the connector.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the connector.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{stringvalidator.LengthBetween(2, 256)},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the staged connector, raised by every change to it.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the connector.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description":       optionalText("Description of the connector."),
			"documentation_url": optionalText("URL of the connector's documentation."),
			"repository_url": schema.StringAttribute{
				MarkdownDescription: "HTTPS or SSH URL of the GitHub repository with the connector's " +
					"`connect.yaml`. A private repository needs its SSH URL.",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"repository_tag": schema.StringAttribute{
				MarkdownDescription: "Git tag of the release to build. Usually set by the commercetools CLI " +
					"on release; see the resource description.",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"integration_types": stringSet("Integration types of the connector, such as `search` or `psp`.",
				stringvalidator.OneOf(enumValues(
					connect.IntegrationTypeTax, connect.IntegrationTypeMarketplace, connect.IntegrationTypeOms,
					connect.IntegrationTypePsp, connect.IntegrationTypePim, connect.IntegrationTypePromotion,
					connect.IntegrationTypeSearch, connect.IntegrationTypeErp, connect.IntegrationTypeCrm,
					connect.IntegrationTypeEmail, connect.IntegrationTypeAnalytics, connect.IntegrationTypeShipping,
					connect.IntegrationTypeGiftcard, connect.IntegrationTypeOther)...)),
			"supported_regions": stringSet("Regions the connector can be deployed in, such as "+
				"`europe-west1.gcp`. Leave out for every region.",
				stringvalidator.OneOf(enumValues(
					connect.RegionEuropeWest1Gcp, connect.RegionUsCentral1Gcp, connect.RegionAustraliaSoutheast1Gcp,
					connect.RegionUsEast2Aws, connect.RegionEuCentral1Aws)...)),
			"private_projects": stringSet("Projects that can deploy the connector, as "+
				"`{region}:{projectKey}`, such as `europe-west1.gcp:my-project`.",
				stringvalidator.RegexMatches(projectIdentifier, "must be {region}:{projectKey}")),
			"publish": schema.BoolAttribute{
				MarkdownDescription: "Publish the connector privately once it is created and whenever " +
					"Terraform changes it, and wait until Connect has published it. Defaults to `false`.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"publish_timeout": schema.StringAttribute{
				MarkdownDescription: "How long to wait for Connect to publish the connector, as a duration " +
					"such as `45m`. Defaults to `30m`.",
				Optional:   true,
				Computed:   true,
				Default:    stringdefault.StaticString("30m"),
				Validators: []validator.String{durationValidator{}},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Status of the staged connector, such as `Draft`, `Processing` or `Published`.",
				Computed:            true,
			},
			"has_changes": schema.BoolAttribute{
				Description: "Whether the staged connector differs from the published one.",
				Computed:    true,
			},
			"published_version": schema.Int64Attribute{
				MarkdownDescription: "Version of the published connector, following the releases " +
					"published with the CLI. Null until the connector is first published.",
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"creator": schema.SingleNestedBlock{
				Description: "Owner of the connector.",
				Validators:  []validator.Object{objectvalidator.IsRequired()},
				Attributes: map[string]schema.Attribute{
					"email": schema.StringAttribute{
						Description: "Contact email address of the creator.",
						Required:    true,
						Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
					},
					"name":        optionalText("Name of the person who owns the connector."),
					"title":       optionalText("Title of the person who owns the connector."),
					"company":     optionalText("Name of the company that owns the connector."),
					"logo_url":    optionalText("URL of a logo representing the creator."),
					"support_url": optionalText("URL of the connector's support website."),
					"no_of_contributors": schema.Int64Attribute{
						Description: "Number of contributors working for the company.",
						Optional:    true,
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
	r.client = data.ConnectorsClient
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan Connector
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft, diags := plan.draft(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	res, err := r.client.Drafts().Post(draft).Execute(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error creating connect connector", processError(err).Error())
		return
	}

	// From here on the connector exists, so it goes into the state whatever
	// the outcome of publishing; a failed one is then tainted and replaced.
	var publishErr error
	if plan.Publish.ValueBool() {
		res, publishErr = r.publish(ctx, res, plan.PublishTimeout)
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
	if publishErr != nil {
		resp.Diagnostics.AddError("Error publishing connect connector", publishErr.Error())
	}
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Connector
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Drafts().WithKey(state.Key.ValueString()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading connect connector", processError(err).Error())
		return
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, state)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Connector
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	actions, diags := plan.updateActions(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	key := state.Key.ValueString()
	res, err := r.client.Drafts().WithKey(key).Get().Execute(ctx)
	if err == nil && len(actions) > 0 {
		res, err = r.client.Drafts().WithKey(key).Post(connect.ConnectorUpdate{
			Version: res.Version,
			Actions: actions,
		}).Execute(ctx)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating connect connector", processError(err).Error())
		return
	}

	// Publish what Terraform changed, or everything when publish was just
	// switched on; a release published by the CLI is left alone.
	var publishErr error
	if plan.Publish.ValueBool() && (len(actions) > 0 || !state.Publish.ValueBool()) {
		res, publishErr = r.publish(ctx, res, plan.PublishTimeout)
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
	if publishErr != nil {
		resp.Diagnostics.AddError("Error publishing connect connector", publishErr.Error())
	}
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Connector
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Drafts().WithKey(state.Key.ValueString()).Delete().Execute(ctx)
	if err != nil && !utils.IsResourceNotFoundError(err) {
		resp.Diagnostics.AddError("Error deleting connect connector", processError(err).Error())
	}
}

// ImportState imports a connector by its key.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("publish"), false)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("publish_timeout"), "30m")...)
}

// setState stores the staged connector with the version of the published
// one, which a deployment refers to.
func (r *Resource) setState(ctx context.Context, state stateSetter, diags *diag.Diagnostics, res *connect.ConnectorStaged, prior Connector) {
	var publishedVersion *int
	published, err := r.client.WithId(res.ID).Get().Execute(ctx)
	switch {
	case err == nil:
		publishedVersion = &published.Version
	case !utils.IsResourceNotFoundError(err):
		diags.AddError("Error reading published connect connector", processError(err).Error())
		return
	}

	current, d := NewConnectorFromNative(ctx, res, publishedVersion, prior)
	diags.Append(d...)
	diags.Append(state.Set(ctx, current)...)
}

type stateSetter interface {
	Set(ctx context.Context, val any) diag.Diagnostics
}

// publish asks Connect to publish the connector privately and returns it
// once Connect has, or an error when publishing failed or did not finish in
// time. Either way it returns the last connector read.
func (r *Resource) publish(ctx context.Context, res *connect.ConnectorStaged, timeout types.String) (*connect.ConnectorStaged, error) {
	next, err := r.client.Drafts().WithId(res.ID).Post(connect.ConnectorUpdate{
		Version: res.Version,
		Actions: []connect.ConnectorUpdateAction{connect.ConnectorPublishAction{Certification: false}},
	}).Execute(ctx)
	if err != nil {
		return res, processError(err)
	}
	res = next

	limit, err := time.ParseDuration(timeout.ValueString())
	if err != nil {
		limit = 30 * time.Minute
	}
	deadline := time.Now().Add(limit)
	for {
		switch connect.ConnectorStatus(res.Status) {
		case connect.ConnectorStatusPublished:
			return res, nil
		case connect.ConnectorStatusFailed:
			return res, fmt.Errorf("connect could not publish %q: %s", res.ID, report(res.PublishingReport))
		}
		if time.Now().After(deadline) {
			return res, fmt.Errorf("timed out after %s; raise publish_timeout to wait longer", limit)
		}
		select {
		case <-ctx.Done():
			return res, ctx.Err()
		case <-time.After(pollInterval):
		}
		next, err := r.client.Drafts().WithId(res.ID).Get().Execute(ctx)
		if err != nil {
			return res, processError(err)
		}
		res = next
	}
}

// report lists the errors in a publishing report.
func report(r *connect.ConnectorReport) string {
	if r == nil {
		return "see the publishing report of the connector"
	}
	var lines []string
	for _, entry := range r.Entries {
		if entry.Type != connect.ConnectorReportEntryTypeError {
			continue
		}
		line := entry.Title
		if entry.Message != nil {
			line += ": " + *entry.Message
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return "see the publishing report of the connector"
	}
	return strings.Join(lines, "; ")
}

func processError(err error) error {
	if retryErr := utils.ProcessRemoteError(err); retryErr != nil && retryErr.Err != nil {
		return retryErr.Err
	}
	return err
}

func enumValues[T ~string](values ...T) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = string(v)
	}
	return out
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
