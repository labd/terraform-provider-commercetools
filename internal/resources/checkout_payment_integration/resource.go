package checkout_payment_integration

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/checkout"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                = &Resource{}
	_ resource.ResourceWithConfigure   = &Resource{}
	_ resource.ResourceWithImportState = &Resource{}
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *checkout.ByProjectKeyRequestBuilder
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_checkout_payment_integration"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	statuses := []string{string(checkout.PaymentIntegrationStatusActive), string(checkout.PaymentIntegrationStatusInactive)}
	replace := []planmodifier.String{stringplanmodifier.RequiresReplace()}
	localized := func(description string) schema.MapAttribute {
		return schema.MapAttribute{
			MarkdownDescription: description,
			ElementType:         types.StringType,
			Optional:            true,
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A payment method offered by a commercetools Checkout application, handled by a " +
			"payment connector deployed with commercetools Connect, such as the Adyen or Stripe connector. " +
			"The connector's deployment is a `commercetools_connect_deployment`; its `type` values come from " +
			"that connector.\n\n" +
			"See also the [Payment Integrations documentation](https://docs.commercetools.com/checkout/payment-integrations-api).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the payment integration.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the payment integration.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(keyPattern, "must match ^[A-Za-z0-9_-]+$"),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the payment integration.",
				Computed:    true,
			},
			"application_id": schema.StringAttribute{
				MarkdownDescription: "ID of the `commercetools_checkout_application` that offers it. Changing " +
					"it replaces the payment integration.",
				Required:      true,
				PlanModifiers: replace,
			},
			"connect_deployment_id": schema.StringAttribute{
				MarkdownDescription: "ID of the `commercetools_connect_deployment` of the payment connector " +
					"that handles it. Changing it replaces the payment integration.",
				Required:      true,
				PlanModifiers: replace,
			},
			"name": schema.StringAttribute{
				Description: "Name of the payment integration in the Merchant Center.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Payment method as the payment service provider and its connector name " +
					"it, such as `card`, `paypal` or `applepay`; `drop-in` for a `DropIn` integration.",
				Required:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"component_type": schema.StringAttribute{
				MarkdownDescription: "`Component`, `DropIn` or `Express`. Checkout offers it only when it " +
					"matches the application's `active_payment_component_type`. Changing it replaces the " +
					"payment integration.",
				Required:      true,
				PlanModifiers: replace,
				Validators: []validator.String{stringvalidator.OneOf(
					string(checkout.PaymentComponentTypeComponent),
					string(checkout.PaymentComponentTypeDropIn),
					string(checkout.PaymentComponentTypeExpress))},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "`Active` or `Inactive`. Defaults to `Inactive`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(checkout.PaymentIntegrationStatusInactive)),
				Validators:          []validator.String{stringvalidator.OneOf(statuses...)},
			},
			"predicate": schema.StringAttribute{
				MarkdownDescription: "JSONata predicate on the cart deciding when it is offered, such as " +
					"`country='NL'`. Only for `Component` integrations.",
				Optional:   true,
				Validators: []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"sort_priority": schema.Int64Attribute{
				Description: "Position among the application's payment methods.",
				Optional:    true,
			},
			"label":           localized("Name shown to the customer, by locale."),
			"description":     localized("Description shown to the customer, by locale."),
			"logo_url":        localized("URL of the logo shown to the customer, by locale."),
			"pay_button_text": localized("Text on the pay button, by locale."),
			"automated_reversal": schema.StringAttribute{
				MarkdownDescription: "`Active` to reverse a payment automatically when the order cannot be " +
					"placed, or `Inactive`.",
				Optional:   true,
				Validators: []validator.String{stringvalidator.OneOf(statuses...)},
			},
			"automated_reversal_predicate": schema.StringAttribute{
				MarkdownDescription: "JSONata predicate on the cart deciding when a payment is reversed " +
					"automatically. Requires `automated_reversal`.",
				Optional: true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
					stringvalidator.AlsoRequires(path.MatchRoot("automated_reversal")),
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
	r.client = data.CheckoutClient
}

func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PaymentIntegration
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft, diags := plan.draft(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var res *checkout.PaymentIntegration
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.PaymentIntegrations().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating checkout payment integration", err.Error())
		return
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PaymentIntegration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.PaymentIntegrations().WithId(state.ID.ValueString()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading checkout payment integration", processError(err).Error())
		return
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, state)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state PaymentIntegration
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
	id := state.ID.ValueString()
	res, err := r.client.PaymentIntegrations().WithId(id).Get().Execute(ctx)
	if err == nil && len(actions) > 0 {
		res, err = r.client.PaymentIntegrations().WithId(id).Post(checkout.PaymentIntegrationUpdateActions{
			Version: res.Version,
			Actions: actions,
		}).Execute(ctx)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating checkout payment integration", processError(err).Error())
		return
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PaymentIntegration
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.PaymentIntegrations().WithId(state.ID.ValueString()).Delete().
			Version(int(state.Version.ValueInt64())).Execute(ctx)
		if utils.IsResourceNotFoundError(err) {
			return nil
		}
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting checkout payment integration", err.Error())
	}
}

// ImportState imports a payment integration by its id.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *Resource) setState(ctx context.Context, state stateSetter, diags *diag.Diagnostics, res *checkout.PaymentIntegration, prior PaymentIntegration) {
	current, d := NewPaymentIntegrationFromNative(ctx, res, prior)
	diags.Append(d...)
	diags.Append(state.Set(ctx, current)...)
}

type stateSetter interface {
	Set(ctx context.Context, val any) diag.Diagnostics
}

func processError(err error) error {
	if retryErr := utils.ProcessRemoteError(err); retryErr != nil && retryErr.Err != nil {
		return retryErr.Err
	}
	return err
}
