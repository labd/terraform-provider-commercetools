package commercetools

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func fieldTypeElementV0(setsAllowed bool) *schema.Resource {
	result := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: func(val any, key string) (warns []string, errs []error) {
				v := val.(string)
				if !setsAllowed && v == "Set" {
					errs = append(errs, fmt.Errorf("sets in another Set are not allowed"))
				}
				return
			},
		},
		"values": {
			Type:     schema.TypeMap,
			Optional: true,
		},
		"localized_value": {
			Type:     schema.TypeList,
			Optional: true,
			Elem:     localizedValueElement(),
		},
		"reference_type_id": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	if setsAllowed {
		result["element_type"] = &schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem:     fieldTypeElementV0(false),
		}
	}

	return &schema.Resource{Schema: result}
}

func resourceTypeResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Identifier for the type (max. 256 characters)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Required:         true,
			},
			"description": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"resource_type_ids": {
				Description: "Defines for which [resources](https://docs.commercetools.com/api/projects/custom-fields#customizable-resources)" +
					" the type is valid",
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"field": {
				Description: "[Field definition](https://docs.commercetools.com/api/projects/types#fielddefinition)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "Describes the [type](https://docs.commercetools.com/api/projects/types#fieldtype)" +
								" of the field",
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem:     fieldTypeElementV0(true),
						},
						"name": {
							Description: "The name of the field.\nThe name must be between two and 36 characters long " +
								"and can contain the ASCII letters A to Z in lowercase or uppercase, digits, " +
								"underscores (_) and the hyphen-minus (-).\nThe name must be unique for a given " +
								"resource type ID. In case there is a field with the same name in another type it has " +
								"to have the same FieldType also",
							Type:     schema.TypeString,
							Required: true,
						},
						"label": {
							Description:      "A human-readable label for the field",
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Required:         true,
						},
						"required": {
							Description: "Whether the field is required to have a value",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"input_hint": {
							Description: "[TextInputHint](https://docs.commercetools.com/api/projects/types#textinputhint)" +
								" Provides a visual representation type for this field. It is only relevant for " +
								"string-based field types like StringType and LocalizedStringType",
							Type:     schema.TypeString,
							Optional: true,
							Default:  platform.TextInputHintSingleLine,
						},
					},
				},
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func migrateTypeStateV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	field, ok := rawState["field"].([]any)
	if !ok {
		return rawState, nil
	}

	// iterate over all fields
	for _, item := range field {
		if field, ok := item.(map[string]any); ok {
			migrateTypeAttributeV0toV1(field)
		}
	}
	return rawState, nil
}

func migrateTypeAttributeV0toV1(attr map[string]any) {
	// check field.type
	itemTypes, ok := attr["type"].([]any)
	if !ok || len(itemTypes) == 1 {
		return
	}

	// it should only contain 1 element, which is an array
	itemType, ok := itemTypes[0].(map[string]any)
	if !ok {
		return
	}

	itemTypeName, ok := itemType["name"].(string)
	if !ok {
		return
	}

	if itemTypeName == "Set" {
		itemTypeElementType, ok := itemType["element_type"].([]any)
		if !ok || len(itemTypeElementType) == 1 {
			return
		}

		itemTypeElementTypeValues, ok := itemTypeElementType[0].(map[string]any)["values"]
		if !ok {
			return
		}

		elementTypeValues, ok := itemTypeElementTypeValues.(map[string]any)
		if !ok {
			return
		}

		// "values" and "value" cannot co exist, so this needs an upgrade
		value := make([]map[string]string, len(elementTypeValues))
		i := 0
		for k, v := range elementTypeValues {
			value[i] = map[string]string{
				"key":   k,
				"label": v.(string),
			}
			i++
		}
		// add "value"
		itemTypeElementType[0].(map[string]any)["value"] = value
		// remove "values"
		delete(itemTypeElementType[0].(map[string]any), "values")

	} else if itemTypeName == "Enum" {
		itemTypeValues, ok := itemType["values"].(map[string]any)
		if !ok {
			return
		}

		// "values" and "value" cannot co exist, so this needs an upgrade
		value := make([]map[string]string, len(itemTypeValues))
		i := 0
		for k, v := range itemTypeValues {
			value[i] = map[string]string{
				"key":   k,
				"label": v.(string),
			}
			i++
		}
		// add "value"
		itemType["value"] = value
		// remove "values"
		delete(itemType, "values")
	}
}
