package checkout_application

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/checkout"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type Application struct {
	ID                         types.String `tfsdk:"id"`
	Key                        types.String `tfsdk:"key"`
	Version                    types.Int64  `tfsdk:"version"`
	Name                       types.String `tfsdk:"name"`
	Description                types.Map    `tfsdk:"description"`
	Status                     types.String `tfsdk:"status"`
	Mode                       types.String `tfsdk:"mode"`
	Countries                  types.Set    `tfsdk:"countries"`
	AllowAllOrigins            types.Bool   `tfsdk:"allow_all_origins"`
	AllowedOrigins             types.Set    `tfsdk:"allowed_origins"`
	LogoURL                    types.String `tfsdk:"logo_url"`
	PaymentReturnURL           types.String `tfsdk:"payment_return_url"`
	ActivePaymentComponentType types.String `tfsdk:"active_payment_component_type"`
	AllowDiscounts             types.Bool   `tfsdk:"allow_discounts"`
	Agreements                 []Agreement  `tfsdk:"agreement"`
}

type Agreement struct {
	Name   types.String `tfsdk:"name"`
	Type   types.String `tfsdk:"type"`
	Status types.String `tfsdk:"status"`
	Text   types.Map    `tfsdk:"text"`
}

// draft is the Application to create. Agreements are left out: a draft takes
// them with ids of their own, so they are set with setAgreements afterwards.
func (a Application) draft(ctx context.Context) (checkout.ApplicationDraft, diag.Diagnostics) {
	var diags diag.Diagnostics
	draft := checkout.ApplicationDraft{
		Key:                    a.Key.ValueString(),
		Status:                 (*checkout.ApplicationStatus)(utils.OptionalString(a.Status)),
		Name:                   a.Name.ValueString(),
		Description:            localized(ctx, a.Description, &diags),
		Logo:                   a.logo(),
		Countries:              strings(ctx, a.Countries, &diags),
		AllowedOrigins:         a.allowedOrigins(ctx, &diags),
		Mode:                   checkout.ApplicationMode(a.Mode.ValueString()),
		PaymentsConfiguration:  a.paymentsConfiguration(),
		DiscountsConfiguration: a.discountsConfiguration(),
	}
	return draft, diags
}

func (a Application) logo() *checkout.ApplicationLogo {
	if a.LogoURL.IsNull() || a.LogoURL.IsUnknown() {
		return nil
	}
	return &checkout.ApplicationLogo{Url: utils.OptionalString(a.LogoURL)}
}

func (a Application) allowedOrigins(ctx context.Context, diags *diag.Diagnostics) checkout.AllowedOrigins {
	return checkout.AllowedOrigins{
		AllowAll: a.AllowAllOrigins.ValueBool(),
		Origins:  strings(ctx, a.AllowedOrigins, diags),
	}
}

func (a Application) paymentsConfiguration() checkout.PaymentsConfiguration {
	return checkout.PaymentsConfiguration{
		PaymentReturnUrl:           utils.OptionalString(a.PaymentReturnURL),
		ActivePaymentComponentType: checkout.PaymentComponentType(a.ActivePaymentComponentType.ValueString()),
	}
}

func (a Application) discountsConfiguration() *checkout.DiscountsConfiguration {
	if a.AllowDiscounts.IsNull() || a.AllowDiscounts.IsUnknown() {
		return nil
	}
	return &checkout.DiscountsConfiguration{AllowDiscounts: a.AllowDiscounts.ValueBool()}
}

func (a Application) agreementDrafts(ctx context.Context, diags *diag.Diagnostics) []checkout.ApplicationAgreementDraft {
	out := make([]checkout.ApplicationAgreementDraft, 0, len(a.Agreements))
	for _, agreement := range a.Agreements {
		text := localized(ctx, agreement.Text, diags)
		if text == nil {
			text = &checkout.LocalizedString{}
		}
		out = append(out, checkout.ApplicationAgreementDraft{
			Name:   agreement.Name.ValueString(),
			Type:   checkout.ApplicationAgreementType(agreement.Type.ValueString()),
			Status: checkout.ApplicationAgreementStatus(agreement.Status.ValueString()),
			Text:   *text,
		})
	}
	return out
}

// updateActions are the actions that change the Application from state into
// a. The nested configurations are each set as a whole.
func (a Application) updateActions(ctx context.Context, state Application) ([]checkout.ApplicationUpdateAction, diag.Diagnostics) {
	var diags diag.Diagnostics
	var actions []checkout.ApplicationUpdateAction

	if !a.Name.Equal(state.Name) {
		actions = append(actions, checkout.SetApplicationNameUpdateAction{Name: a.Name.ValueString()})
	}
	if !a.Description.Equal(state.Description) {
		// An empty object removes the description.
		description := localized(ctx, a.Description, &diags)
		if description == nil {
			description = &checkout.LocalizedString{}
		}
		actions = append(actions, checkout.SetDescriptionUpdateAction{Description: *description})
	}
	if !a.Status.Equal(state.Status) {
		actions = append(actions, checkout.SetApplicationStatusUpdateAction{
			Status: checkout.ApplicationStatus(a.Status.ValueString()),
		})
	}
	if !a.Countries.Equal(state.Countries) {
		actions = append(actions, checkout.SetCountriesUpdateAction{Countries: strings(ctx, a.Countries, &diags)})
	}
	if !a.AllowAllOrigins.Equal(state.AllowAllOrigins) || !a.AllowedOrigins.Equal(state.AllowedOrigins) {
		actions = append(actions, checkout.SetAllowedOriginsUpdateAction{AllowedOrigins: a.allowedOrigins(ctx, &diags)})
	}
	if !a.LogoURL.Equal(state.LogoURL) {
		logo := a.logo()
		if logo == nil {
			logo = &checkout.ApplicationLogo{}
		}
		actions = append(actions, checkout.SetApplicationLogoUpdateAction{Logo: *logo})
	}
	if !a.PaymentReturnURL.Equal(state.PaymentReturnURL) || !a.ActivePaymentComponentType.Equal(state.ActivePaymentComponentType) {
		actions = append(actions, checkout.SetPaymentsConfigurationUpdateAction{PaymentsConfiguration: a.paymentsConfiguration()})
	}
	if !a.AllowDiscounts.Equal(state.AllowDiscounts) {
		actions = append(actions, checkout.SetDiscountsConfigurationUpdateAction{
			DiscountsConfiguration: checkout.DiscountsConfiguration{AllowDiscounts: a.AllowDiscounts.ValueBool()},
		})
	}
	if !slices.EqualFunc(a.Agreements, state.Agreements, Agreement.equal) {
		actions = append(actions, checkout.SetApplicationAgreementsUpdateAction{Agreements: a.agreementDrafts(ctx, &diags)})
	}
	return actions, diags
}

func (a Agreement) equal(b Agreement) bool {
	return a.Name.Equal(b.Name) && a.Type.Equal(b.Type) && a.Status.Equal(b.Status) && a.Text.Equal(b.Text)
}

// NewApplicationFromNative fills the model from the API's answer, in the
// shape of prior, the plan or the state. A value the configuration leaves out
// that the API returns empty stays left out.
func NewApplicationFromNative(ctx context.Context, res *checkout.Application, prior Application) (Application, diag.Diagnostics) {
	var diags diag.Diagnostics
	current := prior
	current.ID = types.StringValue(res.ID)
	current.Key = types.StringValue(res.Key)
	current.Version = types.Int64Value(int64(res.Version))
	current.Name = types.StringValue(res.Name)
	current.Description = fromLocalized(ctx, res.Description, prior.Description, &diags)
	current.Status = types.StringValue(string(checkout.ApplicationStatusInactive))
	if res.Status != nil {
		current.Status = types.StringValue(string(*res.Status))
	}
	current.Mode = types.StringValue(string(res.Mode))
	current.Countries = fromStrings(ctx, res.Countries, types.SetNull(types.StringType), &diags)
	current.AllowAllOrigins = types.BoolValue(res.AllowedOrigins.AllowAll)
	current.AllowedOrigins = fromStrings(ctx, res.AllowedOrigins.Origins, prior.AllowedOrigins, &diags)
	current.LogoURL = types.StringNull()
	if res.Logo != nil && res.Logo.Url != nil && *res.Logo.Url != "" {
		current.LogoURL = types.StringValue(*res.Logo.Url)
	}
	current.PaymentReturnURL = utils.FromOptionalString(res.PaymentsConfiguration.PaymentReturnUrl)
	current.ActivePaymentComponentType = types.StringValue(string(res.PaymentsConfiguration.ActivePaymentComponentType))
	current.AllowDiscounts = types.BoolNull()
	if res.DiscountsConfiguration != nil {
		current.AllowDiscounts = types.BoolValue(res.DiscountsConfiguration.AllowDiscounts)
	}

	current.Agreements = nil
	for _, agreement := range res.Agreements {
		text := agreement.Text
		current.Agreements = append(current.Agreements, Agreement{
			Name:   types.StringValue(agreement.Name),
			Type:   types.StringValue(string(agreement.Type)),
			Status: types.StringValue(string(agreement.Status)),
			Text:   fromLocalized(ctx, &text, types.MapNull(types.StringType), &diags),
		})
	}
	return current, diags
}

func localized(ctx context.Context, m types.Map, diags *diag.Diagnostics) *checkout.LocalizedString {
	if m.IsNull() || m.IsUnknown() {
		return nil
	}
	values := map[string]string{}
	diags.Append(m.ElementsAs(ctx, &values, false)...)
	out := checkout.LocalizedString(values)
	return &out
}

func fromLocalized(ctx context.Context, value *checkout.LocalizedString, prior types.Map, diags *diag.Diagnostics) types.Map {
	if value == nil || (len(*value) == 0 && prior.IsNull()) {
		return types.MapNull(types.StringType)
	}
	m, d := types.MapValueFrom(ctx, types.StringType, map[string]string(*value))
	diags.Append(d...)
	return m
}

func strings(ctx context.Context, set types.Set, diags *diag.Diagnostics) []string {
	out := []string{}
	if set.IsNull() || set.IsUnknown() {
		return out
	}
	diags.Append(set.ElementsAs(ctx, &out, false)...)
	return out
}

func fromStrings(ctx context.Context, values []string, prior types.Set, diags *diag.Diagnostics) types.Set {
	if len(values) == 0 && prior.IsNull() {
		return types.SetNull(types.StringType)
	}
	if values == nil {
		values = []string{}
	}
	set, d := types.SetValueFrom(ctx, types.StringType, values)
	diags.Append(d...)
	return set
}
