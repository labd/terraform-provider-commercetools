package checkout_application

import (
	"context"
	"regexp"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
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
	sdk_resource "github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/labd/commercetools-go-sdk/checkout"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var (
	_ resource.Resource                   = &Resource{}
	_ resource.ResourceWithConfigure      = &Resource{}
	_ resource.ResourceWithImportState    = &Resource{}
	_ resource.ResourceWithValidateConfig = &Resource{}
)

const (
	CompleteFlow = string(checkout.ApplicationModeCompleteFlow)
	PaymentOnly  = string(checkout.ApplicationModePaymentOnly)
)

var keyPattern = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

func NewResource() resource.Resource {
	return &Resource{}
}

type Resource struct {
	client *checkout.ByProjectKeyRequestBuilder
}

func (r *Resource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_checkout_application"
}

func (r *Resource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	statuses := []string{string(checkout.ApplicationStatusActive), string(checkout.ApplicationStatusInactive)}
	localized := func(description string) schema.MapAttribute {
		return schema.MapAttribute{
			MarkdownDescription: description,
			ElementType:         types.StringType,
			Optional:            true,
		}
	}

	resp.Schema = schema.Schema{
		MarkdownDescription: "A commercetools Checkout Application: how Checkout runs for a storefront, the " +
			"countries and origins it serves, and the payment component it shows. Payment methods are added " +
			"with a `commercetools_checkout_payment_integration` for each.\n\n" +
			"See also the [Applications documentation](https://docs.commercetools.com/checkout/applications-api).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key": schema.StringAttribute{
				Description: "User-defined unique identifier of the application.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.LengthBetween(2, 256),
					stringvalidator.RegexMatches(keyPattern, "must match ^[A-Za-z0-9_-]+$"),
				},
			},
			"version": schema.Int64Attribute{
				Description: "Current version of the application.",
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "Name of the application.",
				Required:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"description": localized("Description of the application, by locale."),
			"status": schema.StringAttribute{
				MarkdownDescription: "`Active` or `Inactive`. Defaults to `Inactive`.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(string(checkout.ApplicationStatusInactive)),
				Validators:          []validator.String{stringvalidator.OneOf(statuses...)},
			},
			"mode": schema.StringAttribute{
				MarkdownDescription: "`CompleteFlow` for the whole checkout, or `PaymentOnly` for its payment " +
					"step. Changing it replaces the application.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{stringvalidator.OneOf(CompleteFlow, PaymentOnly)},
			},
			"countries": schema.SetAttribute{
				MarkdownDescription: "Countries the application serves, as ISO 3166-1 alpha-2 codes such as `NL`.",
				ElementType:         types.StringType,
				Required:            true,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
					setvalidator.ValueStringsAre(stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Z]{2}$`),
						"must be an ISO 3166-1 alpha-2 country code")),
				},
			},
			"allow_all_origins": schema.BoolAttribute{
				MarkdownDescription: "Allow every origin to load Checkout. Defaults to `false`; list the " +
					"storefront's origins in `allowed_origins` instead.",
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"allowed_origins": schema.SetAttribute{
				MarkdownDescription: "Origins allowed to load Checkout, such as `https://shop.example.com`.",
				ElementType:         types.StringType,
				Optional:            true,
			},
			"logo_url": schema.StringAttribute{
				MarkdownDescription: "URL of the logo shown in Checkout. Required for `CompleteFlow`.",
				Optional:            true,
				Validators:          []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"payment_return_url": schema.StringAttribute{
				Description: "URL the customer returns to after a payment that redirects away.",
				Optional:    true,
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"active_payment_component_type": schema.StringAttribute{
				MarkdownDescription: "Payment component Checkout shows: `Component`, `DropIn` or `Express`. " +
					"Only payment integrations of this component type are offered.",
				Required: true,
				Validators: []validator.String{stringvalidator.OneOf(
					string(checkout.PaymentComponentTypeComponent),
					string(checkout.PaymentComponentTypeDropIn),
					string(checkout.PaymentComponentTypeExpress))},
			},
			"allow_discounts": schema.BoolAttribute{
				MarkdownDescription: "Let customers enter discount codes. Only for `CompleteFlow`.",
				Optional:            true,
			},
		},
		Blocks: map[string]schema.Block{
			"agreement": schema.ListNestedBlock{
				MarkdownDescription: "An agreement shown to the customer, in the order given. Only for " +
					"`CompleteFlow`.",
				Validators: []validator.List{listvalidator.SizeAtMost(20)},
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the agreement.",
							Required:    true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "`Text`, shown only, or `MandatoryCheckbox`, to be ticked " +
								"before paying.",
							Required: true,
							Validators: []validator.String{stringvalidator.OneOf(
								string(checkout.ApplicationAgreementTypeText),
								string(checkout.ApplicationAgreementTypeMandatoryCheckbox))},
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "`Active` or `Inactive`. Defaults to `Active`.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString(string(checkout.ApplicationAgreementStatusActive)),
							Validators:          []validator.String{stringvalidator.OneOf(statuses...)},
						},
						"text": schema.MapAttribute{
							Description: "Text of the agreement, by locale.",
							ElementType: types.StringType,
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// ValidateConfig rejects what only a CompleteFlow application takes on a
// PaymentOnly one, and a CompleteFlow one without a logo.
func (r *Resource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config Application
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || config.Mode.IsUnknown() {
		return
	}
	switch config.Mode.ValueString() {
	case CompleteFlow:
		if config.LogoURL.IsNull() {
			resp.Diagnostics.AddAttributeError(path.Root("logo_url"), "Missing logo_url",
				"A CompleteFlow application needs a logo_url.")
		}
	case PaymentOnly:
		if !config.AllowDiscounts.IsNull() {
			resp.Diagnostics.AddAttributeError(path.Root("allow_discounts"), "Invalid allow_discounts",
				"Only a CompleteFlow application takes allow_discounts.")
		}
		if len(config.Agreements) > 0 {
			resp.Diagnostics.AddAttributeError(path.Root("agreement"), "Invalid agreement",
				"Only a CompleteFlow application takes agreements.")
		}
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
	var plan Application
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	draft, diags := plan.draft(ctx)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var res *checkout.Application
	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		var err error
		res, err = r.client.Applications().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating checkout application", err.Error())
		return
	}

	if len(plan.Agreements) > 0 {
		agreements := plan.agreementDrafts(ctx, &resp.Diagnostics)
		updated, err := r.update(ctx, res, []checkout.ApplicationUpdateAction{
			checkout.SetApplicationAgreementsUpdateAction{Agreements: agreements},
		})
		if err != nil {
			// The application exists, so it is stored and tainted.
			r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
			resp.Diagnostics.AddError("Error setting checkout application agreements", err.Error())
			return
		}
		res = updated
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
}

func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state Application
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	res, err := r.client.Applications().WithKey(state.Key.ValueString()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading checkout application", processError(err).Error())
		return
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, state)
}

func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state Application
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
	res, err := r.client.Applications().WithKey(state.Key.ValueString()).Get().Execute(ctx)
	if err == nil && len(actions) > 0 {
		res, err = r.update(ctx, res, actions)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error updating checkout application", processError(err).Error())
		return
	}
	r.setState(ctx, &resp.State, &resp.Diagnostics, res, plan)
}

func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state Application
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := sdk_resource.RetryContext(ctx, 20*time.Second, func() *sdk_resource.RetryError {
		_, err := r.client.Applications().WithKey(state.Key.ValueString()).Delete().
			Version(int(state.Version.ValueInt64())).Execute(ctx)
		if utils.IsResourceNotFoundError(err) {
			return nil
		}
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		resp.Diagnostics.AddError("Error deleting checkout application", err.Error())
	}
}

// ImportState imports an application by its key.
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("key"), req, resp)
}

func (r *Resource) update(ctx context.Context, res *checkout.Application, actions []checkout.ApplicationUpdateAction) (*checkout.Application, error) {
	updated, err := r.client.Applications().WithKey(res.Key).Post(checkout.ApplicationUpdateActions{
		Version: res.Version,
		Actions: actions,
	}).Execute(ctx)
	if err != nil {
		return nil, processError(err)
	}
	return updated, nil
}

func (r *Resource) setState(ctx context.Context, state stateSetter, diags *diag.Diagnostics, res *checkout.Application, prior Application) {
	current, d := NewApplicationFromNative(ctx, res, prior)
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
