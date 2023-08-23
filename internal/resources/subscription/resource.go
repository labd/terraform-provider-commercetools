package subscription

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/platform"

	"github.com/labd/terraform-provider-commercetools/internal/customvalidator"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &subscriptionResource{}
	_ resource.ResourceWithConfigure   = &subscriptionResource{}
	_ resource.ResourceWithImportState = &subscriptionResource{}
)

// NewSubscriptionResource is a helper function to simplify the provider implementation.
func NewSubscriptionResource() resource.Resource {
	return &subscriptionResource{}
}

// orderResource is the resource implementation.
type subscriptionResource struct {
	client *platform.ByProjectKeyRequestBuilder
}

// Metadata returns the data source type name.
func (r *subscriptionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subscription"
}

// Schema defines the schema for the data source.
func (r *subscriptionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Subscriptions allow you to be notified of new messages or changes via a Message Queue of your " +
			"choice. Subscriptions are used to trigger an asynchronous background process in response to an event on " +
			"the commercetools platform. Common use cases include sending an Order Confirmation Email, charging a " +
			"Credit Card after the delivery has been made, or synchronizing customer accounts to a Customer " +
			"Relationship Management (CRM) system.\n\n" +
			"See also the [Subscriptions API Documentation](https://docs.commercetools.com/api/projects/subscriptions)",
		Version: 1,
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"key": schema.StringAttribute{
				Description: "Timestamp of the last Terraform update of the order.",
				Optional:    true,
			},
			"version": schema.Int64Attribute{
				Computed: true,
			},
		},
		Blocks: map[string]schema.Block{
			"changes": schema.SetNestedBlock{
				Description: "The change notifications subscribed to",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"resource_type_ids": schema.ListAttribute{
							MarkdownDescription: "[Resource Type ID](https://docs.commercetools.com/api/projects/subscriptions#changesubscription)",
							ElementType:         types.StringType,
							Required:            true,
						},
					},
				},
			},
			"destination": schema.ListNestedBlock{
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf(
									SQS,
									SNS,
									EventBridge,
									EventGrid,
									AzureServiceBus,
									GoogleCloudPubSub,
								),
								customvalidator.DependencyValidator(
									SQS,
									path.MatchRelative().AtParent().AtName("queue_url"),
									path.MatchRelative().AtParent().AtName("region"),
								),
								customvalidator.DependencyValidator(
									SNS,
									path.MatchRelative().AtParent().AtName("topic_arn"),
								),
								customvalidator.DependencyValidator(
									EventBridge,
									path.MatchRelative().AtParent().AtName("account_id"),
									path.MatchRelative().AtParent().AtName("region"),
								),
								customvalidator.DependencyValidator(
									EventGrid,
									path.MatchRelative().AtParent().AtName("access_key"),
									path.MatchRelative().AtParent().AtName("uri"),
								),
								customvalidator.DependencyValidator(
									AzureServiceBus,
									path.MatchRelative().AtParent().AtName("connection_string"),
								),
								customvalidator.DependencyValidator(
									GoogleCloudPubSub,
									path.MatchRelative().AtParent().AtName("project_id"),
									path.MatchRelative().AtParent().AtName("topic"),
								),
							},
						},
						"topic_arn": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"queue_url": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"region": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"account_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"access_key": schema.StringAttribute{
							Optional:   true,
							Validators: []validator.String{
								// TODO Require value if access_secret is set and
								// type is SNS, SQS
							},
							Sensitive: true,
						},
						"access_secret": schema.StringAttribute{
							Optional:   true,
							Validators: []validator.String{
								// TODO Require value if access_key is set and
								// type is SNS, SQS
							},
							Sensitive: true,
						},
						"uri": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"connection_string": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.RegexMatches(
									regexp.MustCompilePOSIX("^Endpoint=sb://"),
									"Connection String should start with Endpoint=sb://",
								),
							},
						},
						"project_id": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"topic": schema.StringAttribute{
							Optional: true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.SizeAtMost(1),
				},
			},
			"format": schema.ListNestedBlock{
				MarkdownDescription: "The [format](https://docs.commercetools.com/api/projects/subscriptions#format) " +
					"in which the payload is delivered",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"type": schema.StringAttribute{
							Optional: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
							Validators: []validator.String{
								stringvalidator.OneOf(
									"Platform", "CloudEvents",
								),
							},
						},
						"cloud_events_version": schema.StringAttribute{
							Description: "For CloudEvents",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.AlsoRequires(
									path.MatchRelative().AtParent().AtName("type"),
								),
							},
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.RequiresReplace(),
							},
						},
					},
				},
				Validators: []validator.List{
					listvalidator.SizeAtMost(1),
				},
			},
			"message": schema.SetNestedBlock{
				Description: "The messages subscribed to",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"resource_type_id": schema.StringAttribute{
							MarkdownDescription: "[Resource Type ID](https://docs.commercetools.com/api/projects/subscriptions#changesubscription)",
							Required:            true,
						},
						"types": schema.ListAttribute{
							MarkdownDescription: "types must contain valid message types for this resource, for example for " +
								"resource type product the message type ProductPublished is valid. If no types of " +
								"messages are given, the subscription is valid for all messages of this resource",
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (r *subscriptionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data := req.ProviderData.(*utils.ProviderData)
	r.client = data.Client
}

func (r *subscriptionResource) UpgradeState(_ context.Context) map[int64]resource.StateUpgrader {
	return map[int64]resource.StateUpgrader{
		0: {
			StateUpgrader: upgradeStateV0,
		},
		1: {
			StateUpgrader: upgradeStateV2,
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *subscriptionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan Subscription
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft := plan.draft()
	var subscription *platform.Subscription
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		subscription, err = r.client.Subscriptions().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating subscription",
			err.Error(),
		)
		return
	}

	current := NewSubscriptionFromNative(subscription)
	current.matchDefaults(plan)
	current.setSecretValues(plan)

	// Set state to fully populated data
	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *subscriptionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state Subscription
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	subscription, err := r.client.Subscriptions().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading subscription",
			"Could not retrieve subscription, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewSubscriptionFromNative(subscription)
	current.matchDefaults(state)
	current.setSecretValues(state)

	// Set refreshed state
	diags = resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *subscriptionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Retrieve values from plan
	var plan Subscription
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Retrieve values from state
	var state Subscription
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := state.updateActions(plan)
	var subscription *platform.Subscription
	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		var err error
		subscription, err = r.client.Subscriptions().WithId(state.ID.ValueString()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating subscription",
			"Could not update subscription, unexpected error: "+err.Error(),
		)
		return
	}

	// Transform response to terraform value and call `setPlanData` with the
	// plan to copy the secrets from the plan since those are returned by
	// commercetools as masked values.
	current := NewSubscriptionFromNative(subscription)
	current.matchDefaults(plan)
	current.setSecretValues(plan)

	diags = resp.State.Set(ctx, current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *subscriptionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state Subscription
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 5*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.Subscriptions().WithId(state.ID.ValueString()).Delete().Version(int(state.Version.ValueInt64())).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting subscription",
			"Could not delete subscription, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *subscriptionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	subscription, err := r.client.Subscriptions().WithId(req.ID).Get().Execute(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error reading subscription",
			"Could not retrieve subscription, unexpected error: "+err.Error(),
		)
		return
	}

	current := NewSubscriptionFromNative(subscription)

	// Set refreshed state
	diags := resp.State.Set(ctx, &current)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
