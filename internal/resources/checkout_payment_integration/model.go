package checkout_payment_integration

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/checkout"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

type PaymentIntegration struct {
	ID                    types.String `tfsdk:"id"`
	Key                   types.String `tfsdk:"key"`
	Version               types.Int64  `tfsdk:"version"`
	ApplicationID         types.String `tfsdk:"application_id"`
	ConnectDeploymentID   types.String `tfsdk:"connect_deployment_id"`
	Name                  types.String `tfsdk:"name"`
	Type                  types.String `tfsdk:"type"`
	ComponentType         types.String `tfsdk:"component_type"`
	Status                types.String `tfsdk:"status"`
	Predicate             types.String `tfsdk:"predicate"`
	SortPriority          types.Int64  `tfsdk:"sort_priority"`
	Label                 types.Map    `tfsdk:"label"`
	Description           types.Map    `tfsdk:"description"`
	LogoURL               types.Map    `tfsdk:"logo_url"`
	PayButtonText         types.Map    `tfsdk:"pay_button_text"`
	AutomatedReversal     types.String `tfsdk:"automated_reversal"`
	AutomatedReversalWhen types.String `tfsdk:"automated_reversal_predicate"`
}

func (p PaymentIntegration) draft(ctx context.Context) (checkout.PaymentIntegrationDraft, diag.Diagnostics) {
	var diags diag.Diagnostics
	draft := checkout.PaymentIntegrationDraft{
		Key:                            utils.OptionalString(p.Key),
		Application:                    checkout.ApplicationReference{ID: p.ApplicationID.ValueString()},
		Status:                         (*checkout.PaymentIntegrationStatus)(utils.OptionalString(p.Status)),
		Type:                           p.Type.ValueString(),
		Name:                           p.Name.ValueString(),
		Predicate:                      utils.OptionalString(p.Predicate),
		ComponentType:                  checkout.PaymentComponentType(p.ComponentType.ValueString()),
		ConnectorDeployment:            checkout.ConnectorDeploymentReference{ID: p.ConnectDeploymentID.ValueString(), TypeId: "deployment"},
		DisplayInfo:                    p.displayInfo(ctx, &diags),
		SortingInfo:                    p.sortingInfo(),
		AutomatedReversalConfiguration: p.automatedReversal(),
	}
	return draft, diags
}

func (p PaymentIntegration) displayInfo(ctx context.Context, diags *diag.Diagnostics) *checkout.DisplayInfo {
	info := checkout.DisplayInfo{
		Label:         localized(ctx, p.Label, diags),
		Description:   localized(ctx, p.Description, diags),
		PayButtonText: localized(ctx, p.PayButtonText, diags),
	}
	if logo := localized(ctx, p.LogoURL, diags); logo != nil {
		url := checkout.LocalizedUrl(*logo)
		info.LogoUrl = &url
	}
	if info.Label == nil && info.Description == nil && info.PayButtonText == nil && info.LogoUrl == nil {
		return nil
	}
	return &info
}

func (p PaymentIntegration) sortingInfo() *checkout.SortingInfo {
	if p.SortPriority.IsNull() || p.SortPriority.IsUnknown() {
		return nil
	}
	return &checkout.SortingInfo{Priority: int(p.SortPriority.ValueInt64())}
}

func (p PaymentIntegration) automatedReversal() *checkout.AutomatedReversalConfiguration {
	if p.AutomatedReversal.IsNull() || p.AutomatedReversal.IsUnknown() {
		return nil
	}
	return &checkout.AutomatedReversalConfiguration{
		Status:    p.AutomatedReversal.ValueString(),
		Predicate: utils.OptionalString(p.AutomatedReversalWhen),
	}
}

// updateActions are the actions that change the payment integration from
// state into p. The display info is set as a whole.
func (p PaymentIntegration) updateActions(ctx context.Context, state PaymentIntegration) ([]checkout.PaymentIntegrationUpdateAction, diag.Diagnostics) {
	var diags diag.Diagnostics
	var actions []checkout.PaymentIntegrationUpdateAction

	if !p.Key.Equal(state.Key) {
		actions = append(actions, checkout.SetKeyUpdateAction{Key: utils.OptionalString(p.Key)})
	}
	if !p.Name.Equal(state.Name) {
		actions = append(actions, checkout.SetNameUpdateAction{Name: p.Name.ValueString()})
	}
	if !p.Type.Equal(state.Type) {
		actions = append(actions, checkout.SetTypeUpdateAction{Type: p.Type.ValueString()})
	}
	if !p.Status.Equal(state.Status) {
		actions = append(actions, checkout.SetStatusUpdateAction{
			Status: checkout.PaymentIntegrationStatus(p.Status.ValueString()),
		})
	}
	if !p.Predicate.Equal(state.Predicate) {
		actions = append(actions, checkout.SetPredicateUpdateAction{Predicate: utils.OptionalString(p.Predicate)})
	}
	if !p.SortPriority.Equal(state.SortPriority) {
		actions = append(actions, checkout.SetSortingInfoUpdateAction{SortingInfo: p.sortingInfo()})
	}
	if !p.Label.Equal(state.Label) || !p.Description.Equal(state.Description) ||
		!p.LogoURL.Equal(state.LogoURL) || !p.PayButtonText.Equal(state.PayButtonText) {
		actions = append(actions, checkout.SetDisplayInfoUpdateAction{DisplayInfo: p.displayInfo(ctx, &diags)})
	}
	if !p.AutomatedReversal.Equal(state.AutomatedReversal) || !p.AutomatedReversalWhen.Equal(state.AutomatedReversalWhen) {
		actions = append(actions, checkout.SetAutomatedReversalConfigurationUpdateAction{
			AutomatedReversalConfiguration: p.automatedReversal(),
		})
	}
	return actions, diags
}

// NewPaymentIntegrationFromNative fills the model from the API's answer, in
// the shape of prior, the plan or the state.
func NewPaymentIntegrationFromNative(ctx context.Context, res *checkout.PaymentIntegration, prior PaymentIntegration) (PaymentIntegration, diag.Diagnostics) {
	var diags diag.Diagnostics
	current := prior
	current.ID = types.StringValue(res.ID)
	current.Key = utils.FromOptionalString(res.Key)
	current.Version = types.Int64Value(int64(res.Version))
	current.ApplicationID = types.StringValue(res.Application.ID)
	current.ConnectDeploymentID = types.StringValue(res.ConnectorDeployment.ID)
	current.Name = types.StringValue(res.Name)
	current.Type = types.StringValue(res.Type)
	current.ComponentType = types.StringValue(string(res.ComponentType))
	current.Status = types.StringValue(string(checkout.PaymentIntegrationStatusInactive))
	if res.Status != nil {
		current.Status = types.StringValue(string(*res.Status))
	}
	current.Predicate = utils.FromOptionalString(res.Predicate)
	current.SortPriority = types.Int64Null()
	if res.SortingInfo != nil {
		current.SortPriority = types.Int64Value(int64(res.SortingInfo.Priority))
	}

	info := checkout.DisplayInfo{}
	if res.DisplayInfo != nil {
		info = *res.DisplayInfo
	}
	current.Label = fromLocalized(ctx, info.Label, &diags)
	current.Description = fromLocalized(ctx, info.Description, &diags)
	current.PayButtonText = fromLocalized(ctx, info.PayButtonText, &diags)
	current.LogoURL = types.MapNull(types.StringType)
	if info.LogoUrl != nil {
		logo := checkout.LocalizedString(*info.LogoUrl)
		current.LogoURL = fromLocalized(ctx, &logo, &diags)
	}

	current.AutomatedReversal = types.StringNull()
	current.AutomatedReversalWhen = types.StringNull()
	if res.AutomatedReversalConfiguration != nil {
		current.AutomatedReversal = types.StringValue(res.AutomatedReversalConfiguration.Status)
		current.AutomatedReversalWhen = utils.FromOptionalString(res.AutomatedReversalConfiguration.Predicate)
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

func fromLocalized(ctx context.Context, value *checkout.LocalizedString, diags *diag.Diagnostics) types.Map {
	if value == nil || len(*value) == 0 {
		return types.MapNull(types.StringType)
	}
	m, d := types.MapValueFrom(ctx, types.StringType, map[string]string(*value))
	diags.Append(d...)
	return m
}
