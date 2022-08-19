package commercetools

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceProductTypeResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"key": {
				Description: "User-specific unique identifier for the product type (max. 256 characters)",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"attribute": {
				Description: "[Product attribute definition](https://docs.commercetools.com/api/projects/productTypes#attributedefinition)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "[AttributeType](https://docs.commercetools.com/api/projects/productTypes#attributetype)",
							Type:        schema.TypeList,
							MaxItems:    1,
							Required:    true,
							Elem:        attributeTypeElementV0(true),
						},
						"name": {
							Description: "The unique name of the attribute used in the API. The name must be between " +
								"two and 256 characters long and can contain the ASCII letters A to Z in lowercase or " +
								"uppercase, digits, underscores (_) and the hyphen-minus (-).\n" +
								"When using the same name for an attribute in two or more product types all fields " +
								"of the AttributeDefinition of this attribute need to be the same across the product " +
								"types, otherwise an AttributeDefinitionAlreadyExists error code will be returned. " +
								"An exception to this are the values of an enum or lenum type and sets thereof",
							Type:     schema.TypeString,
							Required: true,
						},
						"label": {
							Description:      "A human-readable label for the attribute",
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Required:         true,
						},
						"required": {
							Description: "Whether the attribute is required to have a value",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
						"constraint": {
							Description: "Describes how an attribute or a set of attributes should be validated " +
								"across all variants of a product. " +
								"See also [Attribute Constraint](https://docs.commercetools.com/api/projects/productTypes#attributeconstraint-enum)",
							Type:     schema.TypeString,
							Optional: true,
							Default:  platform.AttributeConstraintEnumNone,
							ValidateFunc: func(val any, key string) (warns []string, errs []error) {
								v := val.(string)
								if _, ok := constraintMap[v]; !ok {
									allowedConstraints := []string{}
									for key := range constraintMap {
										allowedConstraints = append(allowedConstraints, key)
									}
									errs = append(errs, fmt.Errorf(
										"unkown attribute constraint '%v'. Possible values are %v", v, allowedConstraints))
								}
								return
							},
						},
						"input_tip": {
							Description: "Additional information about the attribute that aids content managers " +
								"when setting product details",
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Optional:         true,
						},
						"input_hint": {
							Description: "Provides a visual representation type for this attribute. " +
								"only relevant for text-based attribute types like TextType and LocalizableTextType",
							Type:     schema.TypeString,
							Optional: true,
							Default:  platform.TextInputHintSingleLine,
						},
						"searchable": {
							Description: "Whether the attribute's values should generally be activated in product search",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
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

func attributeTypeElementV0(setsAllowed bool) *schema.Resource {
	result := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: func(val any, key string) (warns []string, errs []error) {
				v := val.(string)
				if !setsAllowed && v == "set" {
					errs = append(errs, fmt.Errorf("sets in another Set are not allowed"))
				}
				return
			},
		},
		"value": {
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
		"type_reference": {
			Type:     schema.TypeString,
			Optional: true,
		},
	}

	if setsAllowed {
		result["element_type"] = &schema.Schema{
			Type:     schema.TypeList,
			MaxItems: 1,
			Optional: true,
			Elem:     attributeTypeElement(false),
		}
	}

	return &schema.Resource{Schema: result}
}

func migrateProductTypeStateV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	attr, ok := rawState["attribute"].([]any)
	if !ok {
		return rawState, nil
	}

	// iterate over all attributes
	for _, item := range attr {
		if m, ok := item.(map[string]any); ok {
			migrateProductTypeAttributeV0toV1(m)
		}
	}
	return rawState, nil
}

func migrateProductTypeAttributeV0toV1(attr map[string]any) {
	// check attribute.type
	itemTypes, ok := attr["type"].([]any)
	if !ok || len(itemTypes) != 1 {
		return
	}

	itemType, ok := itemTypes[0].(map[string]any)
	if !ok {
		return
	}

	itemTypeName, ok := itemType["name"].(string)
	if !ok {
		return
	}

	if itemTypeName == "set" {
		itemTypeElementType, ok := itemType["element_type"].([]any)
		if !ok || len(itemTypeElementType) != 1 {
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

	} else if itemTypeName == "enum" {
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
