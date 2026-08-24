package insights_configuration

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func provider(t string, eventTypes ...string) Provider {
	p := Provider{Type: types.StringValue(t)}
	for _, e := range eventTypes {
		p.EventTypes = append(p.EventTypes, types.StringValue(e))
	}
	return p
}

func providerTypes(ps []Provider) []string {
	res := make([]string, 0, len(ps))
	for _, p := range ps {
		res = append(res, p.Type.ValueString())
	}
	return res
}

func TestAlignRestoresApiKey(t *testing.T) {
	// The API strips credentials from its response, so a naive read would wipe
	// the key out of state and show a permanent diff.
	remote := InsightsConfiguration{Providers: []Provider{
		{Type: types.StringValue(NewRelic), ApiKey: types.StringNull()},
	}}
	declared := InsightsConfiguration{Providers: []Provider{
		{Type: types.StringValue(NewRelic), ApiKey: types.StringValue("euNR12345")},
	}}

	remote.alignWith(declared)

	assert.Equal(t, types.StringValue("euNR12345"), remote.Providers[0].ApiKey)
}

func TestAlignRestoresHeaderValuesByName(t *testing.T) {
	remote := InsightsConfiguration{Providers: []Provider{{
		Type: types.StringValue(Otlp),
		Headers: []Header{
			{Name: types.StringValue("api-key"), Value: types.StringNull()},
			{Name: types.StringValue("tenant"), Value: types.StringNull()},
		},
	}}}
	declared := InsightsConfiguration{Providers: []Provider{{
		Type: types.StringValue(Otlp),
		Headers: []Header{
			// Declared in the opposite order to what the API returned.
			{Name: types.StringValue("tenant"), Value: types.StringValue("acme")},
			{Name: types.StringValue("api-key"), Value: types.StringValue("apiKey1234")},
		},
	}}}

	remote.alignWith(declared)

	assert.Equal(t, []Header{
		{Name: types.StringValue("tenant"), Value: types.StringValue("acme")},
		{Name: types.StringValue("api-key"), Value: types.StringValue("apiKey1234")},
	}, remote.Providers[0].Headers)
}

func TestAlignReordersProvidersToDeclaredOrder(t *testing.T) {
	remote := InsightsConfiguration{Providers: []Provider{
		provider(Datadog), provider(NewRelic), provider(Otlp),
	}}
	declared := InsightsConfiguration{Providers: []Provider{
		provider(Otlp), provider(Datadog), provider(NewRelic),
	}}

	remote.alignWith(declared)

	assert.Equal(t, []string{Otlp, Datadog, NewRelic}, providerTypes(remote.Providers))
}

func TestAlignAppendsUndeclaredProviders(t *testing.T) {
	// A provider added outside of Terraform must stay visible so it shows up as
	// a diff rather than being silently dropped.
	remote := InsightsConfiguration{Providers: []Provider{
		provider(NewRelic), provider(Datadog),
	}}
	declared := InsightsConfiguration{Providers: []Provider{provider(NewRelic)}}

	remote.alignWith(declared)

	assert.Equal(t, []string{NewRelic, Datadog}, providerTypes(remote.Providers))
}

func TestAlignDropsDeclaredProviderMissingRemotely(t *testing.T) {
	remote := InsightsConfiguration{Providers: []Provider{provider(NewRelic)}}
	declared := InsightsConfiguration{Providers: []Provider{
		provider(NewRelic), provider(Datadog),
	}}

	remote.alignWith(declared)

	assert.Equal(t, []string{NewRelic}, providerTypes(remote.Providers))
}

func TestAlignKeepsDeclaredEventTypeOrderWhenSetsMatch(t *testing.T) {
	remote := InsightsConfiguration{Providers: []Provider{provider(NewRelic, "Logs", "Metrics")}}
	declared := InsightsConfiguration{Providers: []Provider{provider(NewRelic, "Metrics", "Logs")}}

	remote.alignWith(declared)

	assert.Equal(t, []types.String{
		types.StringValue("Metrics"), types.StringValue("Logs"),
	}, remote.Providers[0].EventTypes)
}

func TestAlignKeepsRemoteEventTypesWhenTheyDiffer(t *testing.T) {
	// A genuine out-of-band change must survive alignment so it is reported.
	remote := InsightsConfiguration{Providers: []Provider{provider(NewRelic, "Logs")}}
	declared := InsightsConfiguration{Providers: []Provider{provider(NewRelic, "Metrics")}}

	remote.alignWith(declared)

	assert.Equal(t, []types.String{types.StringValue("Logs")}, remote.Providers[0].EventTypes)
}

func TestAlignReordersAttributesByName(t *testing.T) {
	attr := func(name, value string) Attribute {
		return Attribute{
			EventType: types.StringValue("Metrics"),
			Name:      types.StringValue(name),
			Value:     types.StringValue(value),
		}
	}
	remote := InsightsConfiguration{AdditionalAttributes: []Attribute{
		attr("zone", "eu"), attr("team", "commercetools"),
	}}
	declared := InsightsConfiguration{AdditionalAttributes: []Attribute{
		attr("team", "commercetools"), attr("zone", "eu"),
	}}

	remote.alignWith(declared)

	assert.Equal(t, []Attribute{
		attr("team", "commercetools"), attr("zone", "eu"),
	}, remote.AdditionalAttributes)
}

func TestAlignWithEmptyPriorStateIsIdentity(t *testing.T) {
	// Import starts from empty state; everything the API returned is kept.
	remote := InsightsConfiguration{Providers: []Provider{
		provider(NewRelic), provider(Datadog),
	}}

	remote.alignWith(InsightsConfiguration{})

	assert.Equal(t, []string{NewRelic, Datadog}, providerTypes(remote.Providers))
}

func TestAlignKeepsRemoteEventTypesWhenOnlyMultiplicityDiffers(t *testing.T) {
	// Same length and every declared value appears remotely, but the lists are
	// still different: a containment check alone would wrongly call these equal.
	remote := InsightsConfiguration{Providers: []Provider{provider(NewRelic, "Logs", "Metrics")}}
	declared := InsightsConfiguration{Providers: []Provider{provider(NewRelic, "Metrics", "Metrics")}}

	remote.alignWith(declared)

	assert.Equal(t, []types.String{
		types.StringValue("Logs"), types.StringValue("Metrics"),
	}, remote.Providers[0].EventTypes)
}
