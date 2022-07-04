package commercetools

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"
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
		CustomizeDiff: customdiff.ValidateChange("attribute", func(ctx context.Context, old, new, meta any) error {
			return resourceProductTypeValidateAttribute(old.([]any), new.([]any))
		}),
	}
}

func attributeTypeElement(setsAllowed bool) *schema.Resource {
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

func resourceProductTypeCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	attributes, err := expandProductTypeAttributeDefinition(d)

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

func flattenProductTypeAttributes(t *platform.ProductType) ([]map[string]any, error) {
	fields := make([]map[string]any, len(t.Attributes))
	for i, attrDef := range t.Attributes {
		attrType, err := flattenProductTypeAttributeType(attrDef.Type, true)
		if err != nil {
			return nil, err
		}
		fields[i] = map[string]any{
			"type":       attrType,
			"name":       attrDef.Name,
			"label":      attrDef.Label,
			"required":   attrDef.IsRequired,
			"input_hint": attrDef.InputHint,
			"constraint": attrDef.AttributeConstraint,
			"searchable": attrDef.IsSearchable,
		}
		if attrDef.InputTip != nil {
			fields[i]["input_tip"] = *attrDef.InputTip
		}
	}
	return fields, nil
}

func resourceProductTypeRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
		d.SetId("")
	} else {
		d.Set("version", ctType.Version)
		d.Set("key", ctType.Key)
		d.Set("name", ctType.Name)
		d.Set("description", ctType.Description)

		if attrs, err := flattenProductTypeAttributes(ctType); err != nil {
			d.Set("attribute", attrs)
		} else {
			return diag.FromErr(err)
		}
	}
	return nil
}

func flattenProductTypeAttributeType(attrType platform.AttributeType, setsAllowed bool) ([]any, error) {
	typeData := make(map[string]any)

	if _, ok := attrType.(platform.AttributeBooleanType); ok {
		typeData["name"] = "boolean"
	} else if _, ok := attrType.(platform.AttributeTextType); ok {
		typeData["name"] = "text"
	} else if _, ok := attrType.(platform.AttributeLocalizableTextType); ok {
		typeData["name"] = "ltext"
	} else if f, ok := attrType.(platform.AttributeEnumType); ok {
		enumValues := make(map[string]any, len(f.Values))
		for _, value := range f.Values {
			enumValues[value.Key] = value.Label
		}
		typeData["name"] = "enum"
		typeData["values"] = enumValues
	} else if f, ok := attrType.(platform.AttributeLocalizedEnumType); ok {
		typeData["name"] = "lenum"
		typeData["localized_value"] = flattenProductTypeLocalizedEnum(f.Values)
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
			elemType, err := flattenProductTypeAttributeType(f.ElementType, false)
			if err != nil {
				return nil, err
			}
			typeData["element_type"] = elemType
		}
	} else {
		return nil, fmt.Errorf("unknown resource Type %T", attrType)
	}

	return []any{typeData}, nil
}

func resourceProductTypeUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
			old.([]any), new.([]any))
		if err != nil {
			return diag.FromErr(err)
		}

		input.Actions = append(input.Actions, attributeChangeActions...)
	}

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.ProductTypes().WithId(d.Id()).Post(input).Execute(ctx)
		return processRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceProductTypeRead(ctx, d, m)
}

func resourceProductTypeDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.ProductTypes().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return processRemoteError(err)
	})
	return diag.FromErr(err)
}

func resourceProductTypeValidateAttribute(old, new []any) error {
	oldLookup := createLookup(old, "name")

	for _, attribute := range new {
		newF := attribute.(map[string]any)
		name := newF["name"].(string)
		oldF, ok := oldLookup[name].(map[string]any)
		if !ok {
			continue
		}

		oldType := firstElementFromSlice(oldF["type"].([]any))
		newType := firstElementFromSlice(newF["type"].([]any))

		oldTypeName := oldType["name"].(string)
		newTypeName := newType["name"].(string)

		if oldTypeName != newTypeName {
			if oldTypeName == "" || newTypeName == "" {
				continue
			}

			return fmt.Errorf(
				"attribute '%s' type changed from %s to %s."+
					" Changing types is not supported;"+
					" please remove the attribute first and re-define it later",
				name, oldTypeName, newTypeName)
		}

		if strings.EqualFold(newTypeName, "Set") {
			oldElement, _ := elementFromSlice(oldType, "element_type")
			newElement, _ := elementFromSlice(newType, "element_type")
			oldElementName := oldElement["name"].(string)
			newElementName := newElement["name"].(string)

			if oldElementName != newElementName {
				return fmt.Errorf(
					"attribute '%s' element type changed from %s to %s."+
						" Changing element types is not supported;"+
						" please remove the attribute first and re-define it later",
					name, oldElementName, newElementName)
			}
		}

		if oldF["required"] != newF["required"] {
			return fmt.Errorf(
				"error on the '%s' attribute: "+
					"Updating the 'required' attribute is not supported."+
					"Consider removing the attribute first and then re-adding it",
				name)
		}
	}
	return nil
}

func resourceProductTypeAttributeChangeActions(oldValues []any, newValues []any) ([]platform.ProductTypeUpdateAction, error) {
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
		newV := value.(map[string]any)
		name := newV["name"].(string)
		oldValue, existingAttr := oldLookup[name]

		var attrDef platform.AttributeDefinition
		if output, err := expandProductTypeAttributeDefinitionItem(newV, false); err == nil {
			attrDef = output.(platform.AttributeDefinition)
		} else {
			return nil, err
		}

		var attrDefDraft platform.AttributeDefinitionDraft
		if output, err := expandProductTypeAttributeDefinitionItem(newV, true); err == nil {
			attrDefDraft = output.(platform.AttributeDefinitionDraft)
		} else {
			return nil, err
		}

		newAttrDefinitions = append(newAttrDefinitions, attrDef)

		if !existingAttr {
			log.Printf("[DEBUG] Attribute added: %s", name)
			actions = append(
				actions,
				platform.ProductTypeAddAttributeDefinitionAction{Attribute: attrDefDraft})
			checkAttributeOrder = false
			continue
		}

		oldV := oldValue.(map[string]any)
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

		newattrType := attrDef.Type
		oldTypes := oldV["type"].([]any)
		var oldattrType map[string]any
		if len(oldTypes) > 0 {
			oldattrType = oldTypes[0].(map[string]any)
		}
		oldEnumKeys := make(map[string]any)
		newEnumKeys := make(map[string]any)

		actions = handleEnumTypeChanges(newattrType, oldattrType, newEnumKeys, actions, name, oldEnumKeys)

		if enumType, ok := newattrType.(platform.AttributeSetType); ok {
			oldElementTypes := oldattrType["element_type"].([]any)
			var myOldattrType map[string]any
			if len(oldElementTypes) > 0 {
				myOldattrType = oldElementTypes[0].(map[string]any)
			}
			actions = handleEnumTypeChanges(enumType.ElementType, myOldattrType, newEnumKeys, actions, name, oldEnumKeys)
			log.Printf("[DEBUG] Set detected: %s", name)
			log.Print(len(myOldattrType))
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
		v := value.(map[string]any)
		oldNames[i] = v["name"].(string)
	}
	for i, value := range newValues {
		v := value.(map[string]any)
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

func handleEnumTypeChanges(newattrType platform.AttributeType, oldattrType map[string]any, newEnumKeys map[string]any, actions []platform.ProductTypeUpdateAction, name string, oldEnumKeys map[string]any) []platform.ProductTypeUpdateAction {
	var (
		oldValues          map[string]any
		oldLocalizedValues []any
		ok                 bool
	)

	if oldattrType != nil {
		if oldValues, ok = oldattrType["values"].(map[string]any); !ok {
			oldValues = make(map[string]any, 0)
		}
		if oldLocalizedValues, ok = oldattrType["localized_value"].([]any); !ok {
			oldLocalizedValues = make([]any, 0)
		}
	}

	if enumType, ok := newattrType.(platform.AttributeEnumType); ok {
		for i, enumValue := range enumType.Values {
			newEnumKeys[enumValue.Key] = enumValue
			if _, ok := oldValues[enumValue.Key]; !ok {
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

	if enumType, ok := newattrType.(platform.AttributeLocalizedEnumType); ok {
		for _, value := range oldLocalizedValues {
			v := value.(map[string]any)
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
				oldEnumValue := oldEnumKeys[enumValue.Key].(map[string]any)
				oldLocalizedLabel := oldEnumValue["label"].(map[string]any)
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

func expandProductTypeAttributeDefinition(d *schema.ResourceData) ([]platform.AttributeDefinitionDraft, error) {
	input := d.Get("attribute").([]any)
	var result []platform.AttributeDefinitionDraft

	for _, raw := range input {
		attrDef, err := expandProductTypeAttributeDefinitionItem(raw.(map[string]any), true)

		if err != nil {
			return nil, err
		}

		result = append(result, attrDef.(platform.AttributeDefinitionDraft))
	}

	return result, nil
}

func expandProductTypeAttributeDefinitionItem(input map[string]any, draft bool) (any, error) {
	attrData, err := elementFromSlice(input, "type")
	if err != nil {
		return nil, err
	}

	attrType, err := expandProductTypeAttributeType(attrData)
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

func expandProductTypeAttributeType(input any) (platform.AttributeType, error) {
	config := input.(map[string]any)
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
		valuesInput, valuesOk := config["values"].(map[string]any)
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
		for _, value := range valuesInput.([]any) {
			v := value.(map[string]any)
			labels := expandLocalizedString(v["label"])

			values = append(values, platform.AttributeLocalizedEnumValue{
				Key:   v["key"].(string),
				Label: labels,
			})
		}
		log.Printf("[DEBUG] expandProductTypeAttributeType localized enum values: %#v", values)
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
		if ref, ok := config["reference_type_id"].(string); ok {
			result := platform.AttributeReferenceType{
				ReferenceTypeId: platform.ReferenceTypeId(ref),
			}
			return result, nil
		}
		return nil, fmt.Errorf("no reference_type_id specified for Reference type")
	case "nested":
		typeReference, typeReferenceOk := config["type_reference"].(string)
		if !typeReferenceOk {
			return nil, fmt.Errorf("no type_reference specified for Nested type")
		}
		return platform.AttributeNestedType{
			TypeReference: platform.ProductTypeReference{ID: typeReference},
		}, nil
	case "set":
		data, err := elementFromSlice(config, "element_type")
		if err != nil {
			return nil, fmt.Errorf("no element_type specified for Set type")
		}

		setAttrType, err := expandProductTypeAttributeType(data)
		if err != nil {
			return nil, err
		}

		return platform.AttributeSetType{
			ElementType: setAttrType,
		}, nil
	}

	return nil, fmt.Errorf("unknown AttributeType %s", typeName)
}

func flattenProductTypeLocalizedEnum(values []platform.AttributeLocalizedEnumValue) []any {
	enumValues := make([]any, len(values))
	for i, value := range values {
		enumValues[i] = map[string]any{
			"key":   value.Key,
			"label": value.Label,
		}
	}
	return enumValues
}
