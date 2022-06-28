package commercetools

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

var constraintMap = map[string]platform.AttributeConstraintEnum{
	"Unique":            platform.AttributeConstraintEnumUnique,
	"CombinationUnique": platform.AttributeConstraintEnumCombinationUnique,
	"SameForAll":        platform.AttributeConstraintEnumSameForAll,
	"None":              platform.AttributeConstraintEnumNone,
}

func resourceProductType() *schema.Resource {
	return &schema.Resource{
		Description: "Product types are used to describe common characteristics, most importantly common custom " +
			"attributes, of many concrete products. Please note: to customize other resources than products, " +
			"please refer to resource_type.\n\n" +
			"See also the [Product Type API Documentation](https://docs.commercetools.com/api/projects/productTypes)",
		CreateContext: resourceProductTypeCreate,
		ReadContext:   resourceProductTypeRead,
		UpdateContext: resourceProductTypeUpdate,
		DeleteContext: resourceProductTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
				Description: "[Product attribute fefinition](https://docs.commercetools.com/api/projects/productTypes#attributedefinition)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "[AttributeType](https://docs.commercetools.com/api/projects/productTypes#attributetype)",
							Type:        schema.TypeList,
							MaxItems:    1,
							Required:    true,
							Elem:        attributeTypeElement(true),
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
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
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
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("attribute", func(ctx context.Context, old, new, meta interface{}) error {
				log.Printf("[DEBUG] Start attribute validation")
				oldLookup := createLookup(old.([]interface{}), "name")
				newV := new.([]interface{})

				for _, field := range newV {
					newF := field.(map[string]interface{})
					name := newF["name"].(string)
					oldF, ok := oldLookup[name].(map[string]interface{})
					if !ok {
						// It means this is a new field, that's ok.
						log.Printf("[DEBUG] Found new attribute: %s", name)
						continue
					}

					log.Printf("[DEBUG] Checking %s", oldF["name"])
					oldType := oldF["type"].([]interface{})[0].(map[string]interface{})
					newType := newF["type"].([]interface{})[0].(map[string]interface{})

					if oldType["name"] != newType["name"] {
						if oldType["name"] != "" || newType["name"] == "" {
							continue
						}
						return fmt.Errorf(
							"field '%s' type changed from %s to %s. Changing types is not supported; please remove the attribute first and re-define it later",
							name, oldType["name"], newType["name"])
					}

					if oldF["required"] != newF["required"] {
						return fmt.Errorf(
							"error on the '%s' attribute: Updating the 'required' attribute is not supported. Consider removing the attribute first and then re-adding it",
							name)
					}
				}
				return nil
			}),
		),
	}
}

func attributeTypeElement(setsAllowed bool) *schema.Resource {
	result := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(string)
				if !setsAllowed && v == "set" {
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

func resourceProductTypeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	attributes, err := resourceProductTypeGetAttributeDefinitions(d)

	if err != nil {
		return diag.FromErr(err)
	}

	draft := platform.ProductTypeDraft{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Attributes:  attributes,
	}

	key := stringRef(d.Get("key"))
	if *key != "" {
		draft.Key = key
	}

	var ctType *platform.ProductType
	err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error

		ctType, err = client.ProductTypes().Post(draft).Execute(ctx)
		return processRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(ctType.ID)
	d.Set("version", ctType.Version)

	return resourceProductTypeRead(ctx, d, m)
}

func resourceProductTypeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Reading product type from commercetools")
	client := getClient(m)

	ctType, err := client.ProductTypes().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	if ctType == nil {
		log.Print("[DEBUG] No product type found")
		d.SetId("")
	} else {
		log.Printf("[DEBUG] Found following product type: %#v", ctType)
		log.Print(stringFormatObject(ctType))

		attributes := make([]map[string]interface{}, len(ctType.Attributes))
		for i, fieldDef := range ctType.Attributes {
			fieldData := make(map[string]interface{})
			log.Printf("[DEBUG] reading field: %s: %#v", fieldDef.Name, fieldDef)
			fieldType, err := resourceProductTypeReadAttributeType(fieldDef.Type, true)
			if err != nil {
				return diag.FromErr(err)
			}

			fieldData["type"] = fieldType
			fieldData["name"] = fieldDef.Name
			fieldData["label"] = fieldDef.Label
			fieldData["required"] = fieldDef.IsRequired
			fieldData["input_hint"] = fieldDef.InputHint
			if fieldDef.InputTip != nil {
				fieldData["input_tip"] = *fieldDef.InputTip
			}
			fieldData["constraint"] = fieldDef.AttributeConstraint
			fieldData["searchable"] = fieldDef.IsSearchable

			attributes[i] = fieldData
		}

		log.Printf("[DEBUG] Created attributes %#v", attributes)
		d.Set("version", ctType.Version)
		d.Set("key", ctType.Key)
		d.Set("name", ctType.Name)
		d.Set("description", ctType.Description)
		err = d.Set("attribute", attributes)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceProductTypeReadAttributeType(attrType platform.AttributeType, setsAllowed bool) ([]interface{}, error) {
	typeData := make(map[string]interface{})

	if _, ok := attrType.(platform.AttributeBooleanType); ok {
		typeData["name"] = "boolean"
	} else if _, ok := attrType.(platform.AttributeTextType); ok {
		typeData["name"] = "text"
	} else if _, ok := attrType.(platform.AttributeLocalizableTextType); ok {
		typeData["name"] = "ltext"
	} else if f, ok := attrType.(platform.AttributeEnumType); ok {
		enumValues := make(map[string]interface{}, len(f.Values))
		for _, value := range f.Values {
			enumValues[value.Key] = value.Label
		}
		typeData["name"] = "enum"
		typeData["values"] = enumValues
	} else if f, ok := attrType.(platform.AttributeLocalizedEnumType); ok {
		typeData["name"] = "lenum"
		typeData["localized_value"] = readAttributeLocalizedEnum(f.Values)
	} else if _, ok := attrType.(platform.AttributeNumberType); ok {
		typeData["name"] = "number"
	} else if _, ok := attrType.(platform.AttributeMoneyType); ok {
		typeData["name"] = "money"
	} else if _, ok := attrType.(platform.AttributeDateType); ok {
		typeData["name"] = "date"
	} else if _, ok := attrType.(platform.AttributeTimeType); ok {
		typeData["name"] = "time"
	} else if _, ok := attrType.(platform.AttributeDateTimeType); ok {
		typeData["name"] = "datetime"
	} else if f, ok := attrType.(platform.AttributeReferenceType); ok {
		typeData["name"] = "reference"
		typeData["reference_type_id"] = f.ReferenceTypeId
	} else if f, ok := attrType.(platform.AttributeNestedType); ok {
		typeData["name"] = "nested"
		typeData["type_reference"] = f.TypeReference.ID
	} else if f, ok := attrType.(platform.AttributeSetType); ok {
		typeData["name"] = "set"
		if setsAllowed {
			elemType, err := resourceProductTypeReadAttributeType(f.ElementType, false)
			if err != nil {
				return nil, err
			}
			typeData["element_type"] = elemType
		}
	} else {
		return nil, fmt.Errorf("unknown resource Type %T", attrType)
	}

	return []interface{}{typeData}, nil
}

func resourceProductTypeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	input := platform.ProductTypeUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.ProductTypeUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.ProductTypeSetKeyAction{Key: &newKey})
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&platform.ProductTypeChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescr := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&platform.ProductTypeChangeDescriptionAction{Description: newDescr})
	}

	if d.HasChange("attribute") {
		old, new := d.GetChange("attribute")
		attributeChangeActions, err := resourceProductTypeAttributeChangeActions(
			old.([]interface{}), new.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}

		input.Actions = append(input.Actions, attributeChangeActions...)
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.ProductTypes().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceProductTypeRead(ctx, d, m)
}

func resourceProductTypeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.ProductTypes().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return processRemoteError(err)
	})
	return diag.FromErr(err)
}

func resourceProductTypeAttributeChangeActions(oldValues []interface{}, newValues []interface{}) ([]platform.ProductTypeUpdateAction, error) {
	oldLookup := createLookup(oldValues, "name")
	newLookup := createLookup(newValues, "name")
	newAttrDefinitions := []platform.AttributeDefinition{}
	actions := []platform.ProductTypeUpdateAction{}
	checkAttributeOrder := true

	for name := range oldLookup {
		if _, ok := newLookup[name]; !ok {
			log.Printf("[DEBUG] Attribute deleted: %s", name)
			actions = append(actions, platform.ProductTypeRemoveAttributeDefinitionAction{Name: name})
			checkAttributeOrder = false
		}
	}

	for _, value := range newValues {
		newV := value.(map[string]interface{})
		name := newV["name"].(string)
		oldValue, existingField := oldLookup[name]

		var attrDef platform.AttributeDefinition
		if output, err := resourceProductTypeGetAttributeDefinition(newV, false); err == nil {
			attrDef = output.(platform.AttributeDefinition)
		} else {
			return nil, err
		}

		var attrDefDraft platform.AttributeDefinitionDraft
		if output, err := resourceProductTypeGetAttributeDefinition(newV, true); err == nil {
			attrDefDraft = output.(platform.AttributeDefinitionDraft)
		} else {
			return nil, err
		}

		newAttrDefinitions = append(newAttrDefinitions, attrDef)

		if !existingField {
			log.Printf("[DEBUG] Attribute added: %s", name)
			actions = append(
				actions,
				platform.ProductTypeAddAttributeDefinitionAction{Attribute: attrDefDraft})
			checkAttributeOrder = false
			continue
		}

		oldV := oldValue.(map[string]interface{})
		if !reflect.DeepEqual(oldV["label"], newV["label"]) {
			actions = append(
				actions,
				platform.ProductTypeChangeLabelAction{
					AttributeName: name, Label: attrDef.Label})
		}
		if oldV["name"] != newV["name"] {
			actions = append(
				actions,
				platform.ProductTypeChangeAttributeNameAction{
					AttributeName: name, NewAttributeName: attrDef.Name})
		}
		if oldV["searchable"] != newV["searchable"] {
			actions = append(
				actions,
				platform.ProductTypeChangeIsSearchableAction{
					AttributeName: name, IsSearchable: attrDef.IsSearchable})
		}
		if oldV["input_hint"] != newV["input_hint"] {
			actions = append(
				actions,
				platform.ProductTypeChangeInputHintAction{
					AttributeName: name, NewValue: attrDef.InputHint})
		}
		if !reflect.DeepEqual(oldV["input_tip"], newV["input_tip"]) {
			actions = append(
				actions,
				platform.ProductTypeSetInputTipAction{
					AttributeName: name,
					InputTip:      attrDef.InputTip,
				})
		}
		if oldV["constraint"] != newV["constraint"] {
			actions = append(
				actions,
				platform.ProductTypeChangeAttributeConstraintAction{
					AttributeName: name,
					NewValue:      platform.AttributeConstraintEnumDraft(attrDef.AttributeConstraint),
				})
		}

		newFieldType := attrDef.Type
		oldFieldType := oldV["type"].([]interface{})[0].(map[string]interface{})
		oldEnumKeys := make(map[string]interface{})
		newEnumKeys := make(map[string]interface{})

		actions = handleEnumTypeChanges(newFieldType, oldFieldType, newEnumKeys, actions, name, oldEnumKeys)

		if enumType, ok := newFieldType.(platform.AttributeSetType); ok {

			myOldFieldType := oldFieldType["element_type"].([]interface{})[0].(map[string]interface{})
			actions = handleEnumTypeChanges(enumType.ElementType, myOldFieldType, newEnumKeys, actions, name, oldEnumKeys)

			log.Printf("[DEBUG] Set detected: %s", name)
			log.Print(len(myOldFieldType))
		}

		removeEnumKeys := []string{}
		for key := range oldEnumKeys {
			if _, ok := newEnumKeys[key]; !ok {
				removeEnumKeys = append(removeEnumKeys, key)
			}
		}

		if len(removeEnumKeys) > 0 {
			actions = append(
				actions,
				platform.ProductTypeRemoveEnumValuesAction{
					AttributeName: name,
					Keys:          removeEnumKeys,
				})
		}

	}

	oldNames := make([]string, len(oldValues))
	newNames := make([]string, len(newValues))

	for i, value := range oldValues {
		v := value.(map[string]interface{})
		oldNames[i] = v["name"].(string)
	}
	for i, value := range newValues {
		v := value.(map[string]interface{})
		newNames[i] = v["name"].(string)
	}

	if checkAttributeOrder && !reflect.DeepEqual(oldNames, newNames) {
		actions = append(
			actions,
			platform.ProductTypeChangeAttributeOrderAction{
				Attributes: newAttrDefinitions,
			})
	}

	log.Printf("[DEBUG] Construction Attribute change actions")

	return actions, nil
}

func handleEnumTypeChanges(newFieldType platform.AttributeType, oldFieldType map[string]interface{}, newEnumKeys map[string]interface{}, actions []platform.ProductTypeUpdateAction, name string, oldEnumKeys map[string]interface{}) []platform.ProductTypeUpdateAction {
	if enumType, ok := newFieldType.(platform.AttributeEnumType); ok {
		oldEnumV := oldFieldType["values"].(map[string]interface{})

		for i, enumValue := range enumType.Values {
			newEnumKeys[enumValue.Key] = enumValue
			if _, ok := oldEnumV[enumValue.Key]; !ok {
				// Key does not appear in old enum values, so we'll add it
				actions = append(
					actions,
					platform.ProductTypeAddPlainEnumValueAction{
						AttributeName: name,
						Value:         enumType.Values[i],
					})
			}
		}

		return actions
		// Action: changePlainEnumValueOrder
		// TODO: Change the order of EnumValues: https://docs.commercetools.com/http-api-projects-productTypes.html#change-the-order-of-enumvalues

	}

	if enumType, ok := newFieldType.(platform.AttributeLocalizedEnumType); ok {
		oldEnumV := oldFieldType["localized_value"].([]interface{})

		for _, value := range oldEnumV {
			v := value.(map[string]interface{})
			oldEnumKeys[v["key"].(string)] = v
		}

		for i, enumValue := range enumType.Values {
			newEnumKeys[enumValue.Key] = enumValue
			if _, ok := oldEnumKeys[enumValue.Key]; !ok {
				// Key does not appear in old enum values, so we'll add it
				actions = append(
					actions,
					platform.ProductTypeAddLocalizedEnumValueAction{
						AttributeName: name,
						Value:         enumType.Values[i],
					})
			} else {
				oldEnumValue := oldEnumKeys[enumValue.Key].(map[string]interface{})
				oldLocalizedLabel := oldEnumValue["label"].(map[string]interface{})
				labelChanged := !localizedStringCompare(enumValue.Label, oldLocalizedLabel)
				if labelChanged {
					actions = append(
						actions,
						platform.ProductTypeChangeLocalizedEnumValueLabelAction{
							AttributeName: name,
							NewValue:      enumType.Values[i],
						})
				}
			}
		}

		return actions
		// Action: changeLocalizedEnumValueOrder
		// TODO: Change the order of LocalizedEnumValues: https://docs.commercetools.com/http-api-projects-productTypes.html#change-the-order-of-localizedenumvalues
	}
	return actions
}

func resourceProductTypeGetAttributeDefinitions(d *schema.ResourceData) ([]platform.AttributeDefinitionDraft, error) {
	input := d.Get("attribute").([]interface{})
	var result []platform.AttributeDefinitionDraft

	for _, raw := range input {
		fieldDef, err := resourceProductTypeGetAttributeDefinition(raw.(map[string]interface{}), true)

		if err != nil {
			return nil, err
		}

		result = append(result, fieldDef.(platform.AttributeDefinitionDraft))
	}

	return result, nil
}

func resourceProductTypeGetAttributeDefinition(input map[string]interface{}, draft bool) (interface{}, error) {
	attrTypes := input["type"].([]interface{})
	attrType, err := getAttributeType(attrTypes[0])
	if err != nil {
		return nil, err
	}

	label := expandLocalizedString(input["label"])
	var inputTip platform.LocalizedString
	if inputTipRaw, ok := input["input_tip"]; ok {
		inputTip = expandLocalizedString(inputTipRaw)
	}

	constraint := platform.AttributeConstraintEnumNone
	constraint, ok := constraintMap[input["constraint"].(string)]
	if !ok {
		constraint = platform.AttributeConstraintEnumNone
	}

	inputHint := platform.TextInputHint(input["input_hint"].(string))
	if draft {
		return platform.AttributeDefinitionDraft{
			Type:                attrType,
			Name:                input["name"].(string),
			Label:               label,
			AttributeConstraint: &constraint,
			IsRequired:          input["required"].(bool),
			IsSearchable:        boolRef(input["searchable"]),
			InputHint:           &inputHint,
			InputTip:            &inputTip,
		}, nil
	}
	return platform.AttributeDefinition{
		Type:                attrType,
		Name:                input["name"].(string),
		Label:               label,
		AttributeConstraint: constraint,
		IsRequired:          input["required"].(bool),
		IsSearchable:        input["searchable"].(bool),
		InputHint:           inputHint,
		InputTip:            &inputTip,
	}, nil
}

func getAttributeType(input interface{}) (platform.AttributeType, error) {
	config := input.(map[string]interface{})
	typeName, ok := config["name"].(string)

	if !ok {
		return nil, fmt.Errorf("no 'name' for type object given")
	}

	switch typeName {
	case "boolean":
		return platform.AttributeBooleanType{}, nil
	case "text":
		return platform.AttributeTextType{}, nil
	case "ltext":
		return platform.AttributeLocalizableTextType{}, nil
	case "enum":
		valuesInput, valuesOk := config["values"].(map[string]interface{})
		if !valuesOk {
			return nil, fmt.Errorf("no values specified for Enum type: %+v", valuesInput)
		}
		var values []platform.AttributePlainEnumValue
		for k, v := range valuesInput {
			values = append(values, platform.AttributePlainEnumValue{
				Key:   k,
				Label: v.(string),
			})
		}
		return platform.AttributeEnumType{Values: values}, nil
	case "lenum":
		valuesInput, valuesOk := config["localized_value"]
		if !valuesOk {
			return nil, fmt.Errorf("no localized_value elements specified for LocalizedEnum type")
		}
		var values []platform.AttributeLocalizedEnumValue
		for _, value := range valuesInput.([]interface{}) {
			v := value.(map[string]interface{})
			labels := expandLocalizedString(v["label"])

			values = append(values, platform.AttributeLocalizedEnumValue{
				Key:   v["key"].(string),
				Label: labels,
			})
		}
		log.Printf("[DEBUG] GetAttributeType localized enum values: %#v", values)
		return platform.AttributeLocalizedEnumType{Values: values}, nil
	case "number":
		return platform.AttributeNumberType{}, nil
	case "money":
		return platform.AttributeMoneyType{}, nil
	case "date":
		return platform.AttributeDateType{}, nil
	case "time":
		return platform.AttributeTimeType{}, nil
	case "datetime":
		return platform.AttributeDateTimeType{}, nil
	case "reference":
		refTypeID, refTypeIDOk := config["reference_type_id"].(string)
		if !refTypeIDOk {
			return nil, fmt.Errorf("no reference_type_id specified for Reference type")
		}
		return platform.AttributeReferenceType{
			ReferenceTypeId: platform.ReferenceTypeId(refTypeID),
		}, nil
	case "nested":
		typeReference, typeReferenceOk := config["type_reference"].(string)
		if !typeReferenceOk {
			return nil, fmt.Errorf("no type_reference specified for Nested type")
		}
		return platform.AttributeNestedType{
			TypeReference: platform.ProductTypeReference{ID: typeReference},
		}, nil
	case "set":
		elementTypes, elementTypesOk := config["element_type"]
		if !elementTypesOk {
			return nil, fmt.Errorf("no element_type specified for Set type")
		}
		elementTypeList := elementTypes.([]interface{})
		if len(elementTypeList) == 0 {
			return nil, fmt.Errorf("no element_type specified for Set type")
		}

		setAttrType, err := getAttributeType(elementTypeList[0])
		if err != nil {
			return nil, err
		}

		return platform.AttributeSetType{
			ElementType: setAttrType,
		}, nil
	}

	return nil, fmt.Errorf("unknown AttributeType %s", typeName)
}

func readAttributeLocalizedEnum(values []platform.AttributeLocalizedEnumValue) []interface{} {
	enumValues := make([]interface{}, len(values))
	for i, value := range values {
		enumValues[i] = map[string]interface{}{
			"key":   value.Key,
			"label": value.Label,
		}
	}
	log.Printf("[DEBUG] readLocalizedEnum values: %#v", enumValues)
	return enumValues
}
