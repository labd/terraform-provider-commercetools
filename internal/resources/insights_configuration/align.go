package insights_configuration

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// alignWith reconciles a freshly-read configuration with the configuration the
// practitioner declared (prior state on a read, the plan on a create/update).
//
// It exists because the API response is not directly usable as Terraform state:
//
//   - Credentials are write-only. A provider's apiKey is absent from the read
//     representation entirely, and OTLP headers come back without their value,
//     so both have to be carried over from what we last sent.
//   - Collection order is not guaranteed to match the order the blocks were
//     written in. Providers are unique by type, attributes by name and headers
//     by name, so each collection is re-sequenced to follow the declared order.
//     Entries that exist remotely but were not declared are appended, which is
//     what surfaces out-of-band changes as a diff instead of hiding them.
func (c *InsightsConfiguration) alignWith(prior InsightsConfiguration) {
	c.Providers = alignProviders(c.Providers, prior.Providers)
	c.AdditionalAttributes = alignAttributes(c.AdditionalAttributes, prior.AdditionalAttributes)
}

func alignProviders(remote, declared []Provider) []Provider {
	res := make([]Provider, 0, len(remote))
	used := make(map[int]bool, len(remote))

	for _, want := range declared {
		for i, got := range remote {
			if used[i] || !got.Type.Equal(want.Type) {
				continue
			}
			used[i] = true
			// Credentials are never returned by the API, so keep what was declared.
			got.ApiKey = want.ApiKey
			got.Headers = alignHeaders(got.Headers, want.Headers)
			got.EventTypes = alignEventTypes(got.EventTypes, want.EventTypes)
			res = append(res, got)
			break
		}
	}

	for i, p := range remote {
		if !used[i] {
			res = append(res, p)
		}
	}

	return res
}

func alignHeaders(remote, declared []Header) []Header {
	res := make([]Header, 0, len(remote))
	used := make(map[int]bool, len(remote))

	for _, want := range declared {
		for i, got := range remote {
			if used[i] || !got.Name.Equal(want.Name) {
				continue
			}
			used[i] = true
			// Header values are write-only, same as apiKey.
			got.Value = want.Value
			res = append(res, got)
			break
		}
	}

	for i, h := range remote {
		if !used[i] {
			res = append(res, h)
		}
	}

	return res
}

func alignAttributes(remote, declared []Attribute) []Attribute {
	res := make([]Attribute, 0, len(remote))
	used := make(map[int]bool, len(remote))

	for _, want := range declared {
		for i, got := range remote {
			if used[i] || !got.Name.Equal(want.Name) {
				continue
			}
			used[i] = true
			res = append(res, got)
			break
		}
	}

	for i, a := range remote {
		if !used[i] {
			res = append(res, a)
		}
	}

	return res
}

// alignEventTypes keeps the declared ordering when both sides hold the same set
// of event types, so that writing ["Metrics", "Logs"] does not read back as a
// diff against ["Logs", "Metrics"].
func alignEventTypes(remote, declared []types.String) []types.String {
	if len(remote) != len(declared) {
		return remote
	}

	// Compared as a multiset rather than with a per-element containment check,
	// so that a repeated value cannot make two different lists look equal.
	counts := make(map[string]int, len(remote))
	for _, e := range remote {
		counts[e.ValueString()]++
	}
	for _, e := range declared {
		counts[e.ValueString()]--
	}
	for _, n := range counts {
		if n != 0 {
			return remote
		}
	}

	return declared
}
