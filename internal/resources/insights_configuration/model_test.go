package insights_configuration

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/labd/commercetools-go-sdk/insights"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

var testModified = time.Date(2024, 12, 12, 12, 1, 3, 0, time.UTC)

func TestNewConfigurationFromNative(t *testing.T) {
	testCases := []struct {
		name string
		n    insights.Provider
		want Provider
	}{
		{
			name: "NewRelic",
			n: insights.NewRelicProvider{
				Active:     true,
				Valid:      utils.Ref(true),
				EventTypes: []insights.EventType{"Metrics"},
				Region:     insights.NewRelicRegionEu,
			},
			want: Provider{
				Type:       types.StringValue(NewRelic),
				Active:     types.BoolValue(true),
				Valid:      types.BoolValue(true),
				EventTypes: []types.String{types.StringValue("Metrics")},
				Region:     types.StringValue("eu"),
				// Never returned by the API.
				ApiKey: types.StringNull(),
			},
		},
		{
			name: "DynatraceSaaS",
			n: insights.DynatraceSaaSProvider{
				Active:        true,
				EventTypes:    []insights.EventType{"Logs", "Metrics"},
				EnvironmentId: "dtMyEnvironment",
			},
			want: Provider{
				Type:          types.StringValue(DynatraceSaaS),
				Active:        types.BoolValue(true),
				Valid:         types.BoolNull(),
				EventTypes:    []types.String{types.StringValue("Logs"), types.StringValue("Metrics")},
				EnvironmentID: types.StringValue("dtMyEnvironment"),
				ApiKey:        types.StringNull(),
			},
		},
		{
			name: "DynatraceActiveGate",
			n: insights.DynatraceActiveGateProvider{
				Active:           true,
				EventTypes:       []insights.EventType{"Metrics"},
				ActiveGateDomain: "my-active-gate.example.com",
				ActiveGatePort:   utils.Ref(float64(9999)),
				EnvironmentId:    "dtMyEnvironment",
			},
			want: Provider{
				Type:             types.StringValue(DynatraceActiveGate),
				Active:           types.BoolValue(true),
				Valid:            types.BoolNull(),
				EventTypes:       []types.String{types.StringValue("Metrics")},
				ActiveGateDomain: types.StringValue("my-active-gate.example.com"),
				ActiveGatePort:   types.Int64Value(9999),
				EnvironmentID:    types.StringValue("dtMyEnvironment"),
				ApiKey:           types.StringNull(),
			},
		},
		{
			name: "Datadog",
			n: insights.DatadogProvider{
				Active:     true,
				EventTypes: []insights.EventType{"Metrics"},
				Site:       insights.DatadogSiteUS1FED,
			},
			want: Provider{
				Type:       types.StringValue(Datadog),
				Active:     types.BoolValue(true),
				Valid:      types.BoolNull(),
				EventTypes: []types.String{types.StringValue("Metrics")},
				Site:       types.StringValue("US1-FED"),
				ApiKey:     types.StringNull(),
			},
		},
		{
			name: "Otlp",
			n: insights.OtlpProvider{
				Active:     true,
				EventTypes: []insights.EventType{"Metrics"},
				Endpoint:   "https://myotlpprovider.com:14317",
				Headers:    []insights.OtlpProviderHeader{{Name: "api-key"}},
			},
			want: Provider{
				Type:       types.StringValue(Otlp),
				Active:     types.BoolValue(true),
				Valid:      types.BoolNull(),
				EventTypes: []types.String{types.StringValue("Metrics")},
				Endpoint:   types.StringValue("https://myotlpprovider.com:14317"),
				ApiKey:     types.StringNull(),
				Headers: []Header{{
					Name: types.StringValue("api-key"),
					// Header values are never returned by the API.
					Value: types.StringNull(),
				}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			native := &insights.ProjectConfiguration{
				Active:         true,
				ProjectKey:     "my-project",
				LastModifiedAt: testModified,
				Providers:      []insights.Provider{tc.n},
				AdditionalAttributes: []insights.Attribute{
					insights.StringAttribute{EventType: "Metrics", Name: "team", Value: "commercetools"},
				},
			}

			got := NewConfigurationFromNative(native)

			assert.Equal(t, types.StringValue("my-project"), got.ID)
			assert.Equal(t, types.StringValue("my-project"), got.ProjectKey)
			assert.Equal(t, types.BoolValue(true), got.Active)
			assert.Equal(t, types.StringValue("2024-12-12T12:01:03Z"), got.LastModifiedAt)
			assert.Equal(t, []Provider{tc.want}, got.Providers)
			assert.Equal(t, []Attribute{{
				EventType: types.StringValue("Metrics"),
				Name:      types.StringValue("team"),
				Value:     types.StringValue("commercetools"),
			}}, got.AdditionalAttributes)
		})
	}
}

func TestDraft(t *testing.T) {
	testCases := []struct {
		name string
		p    Provider
		want insights.ProviderDraft
	}{
		{
			name: "NewRelic",
			p: Provider{
				Type:       types.StringValue(NewRelic),
				Active:     types.BoolValue(true),
				EventTypes: []types.String{types.StringValue("Metrics")},
				Region:     types.StringValue("eu"),
				ApiKey:     types.StringValue("euNR12345"),
			},
			want: insights.NewRelicProviderDraft{
				Active:     utils.Ref(true),
				EventTypes: []insights.EventType{"Metrics"},
				Region:     insights.NewRelicRegionEu,
				ApiKey:     utils.Ref("euNR12345"),
			},
		},
		{
			name: "DynatraceActiveGate with port",
			p: Provider{
				Type:             types.StringValue(DynatraceActiveGate),
				Active:           types.BoolNull(),
				EventTypes:       []types.String{types.StringValue("Logs")},
				ActiveGateDomain: types.StringValue("my-active-gate.example.com"),
				ActiveGatePort:   types.Int64Value(9999),
				EnvironmentID:    types.StringValue("dtMyEnvironment"),
				ApiKey:           types.StringValue("dtabc1234"),
			},
			want: insights.DynatraceActiveGateProviderDraft{
				EventTypes:       []insights.EventType{"Logs"},
				ActiveGateDomain: "my-active-gate.example.com",
				ActiveGatePort:   utils.Ref(float64(9999)),
				EnvironmentId:    "dtMyEnvironment",
				ApiKey:           utils.Ref("dtabc1234"),
			},
		},
		{
			name: "OtlpHttp with headers",
			p: Provider{
				Type:       types.StringValue(OtlpHttp),
				EventTypes: []types.String{types.StringValue("Metrics")},
				Endpoint:   types.StringValue("https://myotlpprovider.com:14318"),
				Headers: []Header{{
					Name:  types.StringValue("api-key"),
					Value: types.StringValue("apiKey1234"),
				}},
			},
			want: insights.OtlpHttpProviderDraft{
				EventTypes: []insights.EventType{"Metrics"},
				Endpoint:   "https://myotlpprovider.com:14318",
				Headers: []insights.OtlpProviderHeaderDraft{
					{Name: "api-key", Value: "apiKey1234"},
				},
			},
		},
		{
			name: "omits api_key when not configured so the existing key is kept",
			p: Provider{
				Type:       types.StringValue(Datadog),
				EventTypes: []types.String{types.StringValue("Metrics")},
				Site:       types.StringValue("US1"),
				ApiKey:     types.StringNull(),
			},
			want: insights.DatadogProviderDraft{
				EventTypes: []insights.EventType{"Metrics"},
				Site:       insights.DatadogSiteUS1,
				ApiKey:     nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := InsightsConfiguration{
				Active:    types.BoolValue(true),
				Providers: []Provider{tc.p},
			}

			got := c.draft()

			assert.Equal(t, utils.Ref(true), got.Active)
			assert.Equal(t, []insights.ProviderDraft{tc.want}, got.Providers)
		})
	}
}

func TestDraftOmitsUnknownActive(t *testing.T) {
	c := InsightsConfiguration{Active: types.BoolUnknown()}
	assert.Nil(t, c.draft().Active)
}
