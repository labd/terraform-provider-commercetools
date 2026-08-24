package insights_configuration

import (
	"time"

	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/insights"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

// Supported APM provider types.
const (
	NewRelic            = "NewRelic"
	DynatraceSaaS       = "DynatraceSaaS"
	DynatraceActiveGate = "DynatraceActiveGate"
	Datadog             = "Datadog"
	Otlp                = "Otlp"
	OtlpHttp            = "OtlpHttp"
)

// InsightsConfiguration is the main resource schema data. The Platform Insights
// configuration is a singleton per Project, so it has no server-side identifier
// of its own; the Project key doubles as the Terraform ID.
type InsightsConfiguration struct {
	ID                   types.String `tfsdk:"id"`
	ProjectKey           types.String `tfsdk:"project_key"`
	Active               types.Bool   `tfsdk:"active"`
	LastModifiedAt       types.String `tfsdk:"last_modified_at"`
	Providers            []Provider   `tfsdk:"apm_provider"`
	AdditionalAttributes []Attribute  `tfsdk:"additional_attribute"`
}

// Provider is a single APM provider entry. The commercetools API models these
// as a discriminated union on `type`; following the convention used by the
// subscription resource they are flattened into one block where the fields that
// do not apply to the chosen type are left null.
type Provider struct {
	Type             types.String   `tfsdk:"type"`
	Active           types.Bool     `tfsdk:"active"`
	Valid            types.Bool     `tfsdk:"valid"`
	EventTypes       []types.String `tfsdk:"event_types"`
	Region           types.String   `tfsdk:"region"`
	Site             types.String   `tfsdk:"site"`
	EnvironmentID    types.String   `tfsdk:"environment_id"`
	ActiveGateDomain types.String   `tfsdk:"active_gate_domain"`
	ActiveGatePort   types.Int64    `tfsdk:"active_gate_port"`
	Endpoint         types.String   `tfsdk:"endpoint"`
	ApiKey           types.String   `tfsdk:"api_key"`
	Headers          []Header       `tfsdk:"header"`
}

// Header is a request header sent to an OTLP provider.
type Header struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

// Attribute is an additional attribute added to exported events.
type Attribute struct {
	EventType types.String `tfsdk:"event_type"`
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
}

// NewConfigurationFromNative converts an API response into the resource model.
//
// Note that the API never returns provider credentials: the read representation
// of a provider has no apiKey field at all, and OTLP headers come back with
// only their name. Those land as null here and are restored from prior state or
// the plan by alignWith.
func NewConfigurationFromNative(n *insights.ProjectConfiguration) InsightsConfiguration {
	res := InsightsConfiguration{
		ID:                   types.StringValue(n.ProjectKey),
		ProjectKey:           types.StringValue(n.ProjectKey),
		Active:               types.BoolValue(n.Active),
		LastModifiedAt:       types.StringValue(n.LastModifiedAt.Format(time.RFC3339)),
		Providers:            make([]Provider, 0, len(n.Providers)),
		AdditionalAttributes: make([]Attribute, 0, len(n.AdditionalAttributes)),
	}

	for _, p := range n.Providers {
		res.Providers = append(res.Providers, newProviderFromNative(p))
	}

	for _, a := range n.AdditionalAttributes {
		if attr, ok := a.(insights.StringAttribute); ok {
			res.AdditionalAttributes = append(res.AdditionalAttributes, Attribute{
				EventType: types.StringValue(string(attr.EventType)),
				Name:      types.StringValue(attr.Name),
				Value:     types.StringValue(attr.Value),
			})
		}
	}

	return res
}

func newProviderFromNative(n insights.Provider) Provider {
	// Fields that do not apply to the provider type at hand stay null.
	p := Provider{
		Type:             types.StringNull(),
		Region:           types.StringNull(),
		Site:             types.StringNull(),
		EnvironmentID:    types.StringNull(),
		ActiveGateDomain: types.StringNull(),
		ActiveGatePort:   types.Int64Null(),
		Endpoint:         types.StringNull(),
		ApiKey:           types.StringNull(),
	}

	switch v := n.(type) {
	case insights.NewRelicProvider:
		p.Type = types.StringValue(NewRelic)
		p.Active = types.BoolValue(v.Active)
		p.Valid = utils.FromOptionalBool(v.Valid)
		p.EventTypes = eventTypesToTerraform(v.EventTypes)
		p.Region = types.StringValue(string(v.Region))

	case insights.DynatraceSaaSProvider:
		p.Type = types.StringValue(DynatraceSaaS)
		p.Active = types.BoolValue(v.Active)
		p.Valid = utils.FromOptionalBool(v.Valid)
		p.EventTypes = eventTypesToTerraform(v.EventTypes)
		p.EnvironmentID = types.StringValue(v.EnvironmentId)

	case insights.DynatraceActiveGateProvider:
		p.Type = types.StringValue(DynatraceActiveGate)
		p.Active = types.BoolValue(v.Active)
		p.Valid = utils.FromOptionalBool(v.Valid)
		p.EventTypes = eventTypesToTerraform(v.EventTypes)
		p.ActiveGateDomain = types.StringValue(v.ActiveGateDomain)
		p.EnvironmentID = types.StringValue(v.EnvironmentId)
		if v.ActiveGatePort != nil {
			p.ActiveGatePort = types.Int64Value(int64(*v.ActiveGatePort))
		}

	case insights.DatadogProvider:
		p.Type = types.StringValue(Datadog)
		p.Active = types.BoolValue(v.Active)
		p.Valid = utils.FromOptionalBool(v.Valid)
		p.EventTypes = eventTypesToTerraform(v.EventTypes)
		p.Site = types.StringValue(string(v.Site))

	case insights.OtlpProvider:
		p.Type = types.StringValue(Otlp)
		p.Active = types.BoolValue(v.Active)
		p.Valid = utils.FromOptionalBool(v.Valid)
		p.EventTypes = eventTypesToTerraform(v.EventTypes)
		p.Endpoint = types.StringValue(v.Endpoint)
		p.Headers = headersFromNative(v.Headers)

	case insights.OtlpHttpProvider:
		p.Type = types.StringValue(OtlpHttp)
		p.Active = types.BoolValue(v.Active)
		p.Valid = utils.FromOptionalBool(v.Valid)
		p.EventTypes = eventTypesToTerraform(v.EventTypes)
		p.Endpoint = types.StringValue(v.Endpoint)
		p.Headers = headersFromNative(v.Headers)
	}

	return p
}

// headersFromNative converts read headers. The API only returns header names,
// so values are null until alignWith restores them.
func headersFromNative(n []insights.OtlpProviderHeader) []Header {
	res := make([]Header, 0, len(n))
	for _, h := range n {
		res = append(res, Header{
			Name:  types.StringValue(h.Name),
			Value: types.StringNull(),
		})
	}
	return res
}

func eventTypesToTerraform(n []insights.EventType) []types.String {
	return pie.Map(n, func(e insights.EventType) types.String {
		return types.StringValue(string(e))
	})
}

func eventTypesToNative(n []types.String) []insights.EventType {
	return pie.Map(n, func(e types.String) insights.EventType {
		return insights.EventType(e.ValueString())
	})
}

// draft builds the payload for a create-or-replace (PUT) request.
func (c InsightsConfiguration) draft() insights.ProjectConfigurationDraft {
	draft := insights.ProjectConfigurationDraft{
		Providers:            make([]insights.ProviderDraft, 0, len(c.Providers)),
		AdditionalAttributes: make([]insights.AttributeDraft, 0, len(c.AdditionalAttributes)),
	}

	if !c.Active.IsNull() && !c.Active.IsUnknown() {
		draft.Active = utils.Ref(c.Active.ValueBool())
	}

	for _, p := range c.Providers {
		draft.Providers = append(draft.Providers, p.toNative())
	}

	for _, a := range c.AdditionalAttributes {
		draft.AdditionalAttributes = append(draft.AdditionalAttributes, insights.StringAttributeDraft{
			EventType: insights.EventType(a.EventType.ValueString()),
			Name:      a.Name.ValueString(),
			Value:     a.Value.ValueString(),
		})
	}

	return draft
}

func (p Provider) toNative() insights.ProviderDraft {
	var active *bool
	if !p.Active.IsNull() && !p.Active.IsUnknown() {
		active = utils.Ref(p.Active.ValueBool())
	}
	eventTypes := eventTypesToNative(p.EventTypes)

	switch p.Type.ValueString() {
	case NewRelic:
		return insights.NewRelicProviderDraft{
			Active:     active,
			EventTypes: eventTypes,
			Region:     insights.NewRelicRegion(p.Region.ValueString()),
			ApiKey:     utils.OptionalString(p.ApiKey),
		}

	case DynatraceSaaS:
		return insights.DynatraceSaaSProviderDraft{
			Active:        active,
			EventTypes:    eventTypes,
			EnvironmentId: p.EnvironmentID.ValueString(),
			ApiKey:        utils.OptionalString(p.ApiKey),
		}

	case DynatraceActiveGate:
		draft := insights.DynatraceActiveGateProviderDraft{
			Active:           active,
			EventTypes:       eventTypes,
			ActiveGateDomain: p.ActiveGateDomain.ValueString(),
			EnvironmentId:    p.EnvironmentID.ValueString(),
			ApiKey:           utils.OptionalString(p.ApiKey),
		}
		if !p.ActiveGatePort.IsNull() && !p.ActiveGatePort.IsUnknown() {
			draft.ActiveGatePort = utils.Ref(float64(p.ActiveGatePort.ValueInt64()))
		}
		return draft

	case Datadog:
		return insights.DatadogProviderDraft{
			Active:     active,
			EventTypes: eventTypes,
			Site:       insights.DatadogSite(p.Site.ValueString()),
			ApiKey:     utils.OptionalString(p.ApiKey),
		}

	case Otlp:
		return insights.OtlpProviderDraft{
			Active:     active,
			EventTypes: eventTypes,
			Endpoint:   p.Endpoint.ValueString(),
			Headers:    p.headersToNative(),
		}

	case OtlpHttp:
		return insights.OtlpHttpProviderDraft{
			Active:     active,
			EventTypes: eventTypes,
			Endpoint:   p.Endpoint.ValueString(),
			Headers:    p.headersToNative(),
		}
	}

	return nil
}

func (p Provider) headersToNative() []insights.OtlpProviderHeaderDraft {
	if len(p.Headers) == 0 {
		return nil
	}
	res := make([]insights.OtlpProviderHeaderDraft, 0, len(p.Headers))
	for _, h := range p.Headers {
		res = append(res, insights.OtlpProviderHeaderDraft{
			Name:  h.Name.ValueString(),
			Value: h.Value.ValueString(),
		})
	}
	return res
}
