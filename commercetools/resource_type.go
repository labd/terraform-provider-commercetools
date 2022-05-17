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

func resourceType() *schema.Resource {
	return &schema.Resource{
		Description: "Types define custom fields that are used to enhance resources as you need. Use Types to model " +
			"your own CustomFields on resources, like Category and Customer.\n\n" +
			"In case you want to customize products, please use product types instead that serve a similar purpose, " +
			"but tailored to products.\n\n" +
			"See also the [Types Api Documentation](https://docs.commercetools.com/api/projects/types)",
		CreateContext: resourceTypeCreate,
		ReadContext:   resourceTypeRead,
		UpdateContext: resourceTypeUpdate,
		DeleteContext: resourceTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
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
							Elem:     fieldTypeElement(true),
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
		CustomizeDiff: customdiff.All(
			customdiff.ValidateChange("field", func(ctx context.Context, old, new, meta interface{}) error {
				log.Printf("[DEBUG] Start field validation")
				oldLookup := createLookup(old.([]interface{}), "name")
				newV := new.([]interface{})

				for _, field := range newV {
					newF := field.(map[string]interface{})
					name := newF["name"].(string)
					oldF, ok := oldLookup[name].(map[string]interface{})
					if !ok {
						// It means this is a new field, that's ok.
						log.Printf("[DEBUG] Found new field: %s", name)
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
							"field '%s' type changed from %s to %s. Changing types is not supported; please remove the field first and re-define it later",
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

func localizedValueElement() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"label": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Required:         true,
			},
		},
	}
}

func fieldTypeElement(setsAllowed bool) *schema.Resource {
	result := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
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
		// Or, alternatively, we could go with the following
		// to have it more consistent with localized_value.
		// However, this is the difference between:
		// |	values = {
		// |		value1 = "Value 1"
		// |		value2 = "Value 2"
		// |	}
		//  and
		// |	value {
		// |		key = "value1"
		// |		label = "Value 1"
		// |	}
		// |	value {
		// |		key = "value2"
		// |		label = "Value 2"
		// |	}
		// "value": {
		// 	Type:     schema.TypeSet,
		// 	Optional: true,
		// 	Elem: &schema.Resource{
		// 		Schema: map[string]*schema.Schema{
		// 			"key": {
		// 				Type:     schema.TypeString,
		// 				Required: true,
		// 			},
		// 			"label": {
		// 				Type:     schema.TypeString,
		// 				Required: true,
		// 			},
		// 		},
		// 	},
		// },
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
			Elem:     fieldTypeElement(false),
		}
	}

	return &schema.Resource{Schema: result}
}

func resourceTypeCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	var ctType *platform.Type

	name := unmarshallLocalizedString(d.Get("name"))
	description := unmarshallLocalizedString(d.Get("description"))

	resourceTypeIds := []platform.ResourceTypeId{}
	for _, item := range expandStringArray(d.Get("resource_type_ids").([]interface{})) {
		resourceTypeIds = append(resourceTypeIds, platform.ResourceTypeId(item))

	}
	fields, err := resourceTypeGetFieldDefinitions(d)

	if err != nil {
		return diag.FromErr(err)
	}

	draft := platform.TypeDraft{
		Key:              d.Get("key").(string),
		Name:             name,
		Description:      &description,
		ResourceTypeIds:  resourceTypeIds,
		FieldDefinitions: fields,
	}

	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error

		ctType, err = client.Types().Post(draft).Execute(ctx)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if ctType == nil {
		return diag.Errorf("No type created?")
	}

	d.SetId(ctType.ID)
	d.Set("version", ctType.Version)

	return resourceTypeRead(ctx, d, m)
}

func resourceTypeRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Print("[DEBUG] Reading type from commercetools")
	client := getClient(m)

	ctType, err := client.Types().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	if ctType == nil {
		log.Print("[DEBUG] No type found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following type:")
		log.Print(stringFormatObject(ctType))

		d.Set("version", ctType.Version)
		d.Set("key", ctType.Key)
		d.Set("name", ctType.Name)
		if ctType.Description != nil {
			d.Set("description", ctType.Description)
		}
		d.Set("resource_type_ids", ctType.ResourceTypeIds)

		if fields, err := marshallTypeFields(ctType); err == nil {
			d.Set("field", fields)
		} else {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceTypeUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	input := platform.TypeUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.TypeUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.TypeChangeKeyAction{Key: newKey})
	}

	if d.HasChange("name") {
		newName := unmarshallLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.TypeChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := unmarshallLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.TypeSetDescriptionAction{
				Description: &newDescription})
	}

	if d.HasChange("field") {
		old, new := d.GetChange("field")
		fieldChangeActions, err := resourceTypeFieldChangeActions(old.([]interface{}), new.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		input.Actions = append(input.Actions, fieldChangeActions...)
	}
	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err := client.Types().WithId(d.Id()).Post(input).Execute(ctx)
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return diag.FromErr(err)
	}

	return resourceTypeRead(ctx, d, m)
}

// Generate a list of actions needed for updating the fields value in
// commercetools so that it matches the terraform file
func resourceTypeFieldChangeActions(oldValues []interface{}, newValues []interface{}) ([]platform.TypeUpdateAction, error) {
	oldLookup := createLookup(oldValues, "name")
	newLookup := createLookup(newValues, "name")
	actions := []platform.TypeUpdateAction{}
	checkAttributeOrder := true

	log.Printf("[DEBUG] Construction Field change actions")

	// Check if we have fields which are removed and generate the corresponding
	// remove field actions
	for name := range oldLookup {
		if _, ok := newLookup[name]; !ok {
			log.Printf("[DEBUG] Field deleted: %s", name)
			actions = append(actions, platform.TypeRemoveFieldDefinitionAction{FieldName: name})
			// checkAttributeOrder = false
		}
	}

	for i := range newValues {
		newV := newValues[i].(map[string]interface{})
		name := newV["name"].(string)
		oldValue, existingField := oldLookup[name]

		fieldDef, err := resourceTypeGetFieldDefinition(newV)
		if err != nil {
			return nil, err
		}

		// A new field is added. Create the update action skip the rest of the
		// loop since there cannot be any change if the field didn't exist yet.
		if !existingField {
			log.Printf("[DEBUG] Field added: %s", name)
			actions = append(
				actions,
				platform.TypeAddFieldDefinitionAction{FieldDefinition: *fieldDef})
			// checkAttributeOrder = false
			continue
		}

		// Check if we need to update the field label
		oldV := oldValue.(map[string]interface{})
		if !reflect.DeepEqual(oldV["label"], newV["label"]) {
			newLabel := unmarshallLocalizedString(newV["label"])
			actions = append(
				actions,
				platform.TypeChangeLabelAction{FieldName: name, Label: newLabel})
		}

		// Update the input hint if this is changed
		if !reflect.DeepEqual(oldV["input_hint"], newV["input_hint"]) {
			var newInputHint platform.TypeTextInputHint
			switch newV["input_hint"].(string) {
			case "SingleLine":
				newInputHint = platform.TypeTextInputHintSingleLine
			case "MultiLine":
				newInputHint = platform.TypeTextInputHintMultiLine
			}

			actions = append(
				actions,
				platform.TypeChangeInputHintAction{FieldName: name, InputHint: newInputHint})
		}

		newFieldType := fieldDef.Type
		oldFieldType := oldV["type"].([]interface{})[0].(map[string]interface{})

		if enumType, ok := newFieldType.(platform.CustomFieldSetType); ok {

			myOldFieldType := oldFieldType["element_type"].([]interface{})[0].(map[string]interface{})
			actions = resourceTypeHandleEnumTypeChanges(enumType.ElementType, myOldFieldType, actions, name)

			log.Printf("[DEBUG] Set detected: %s", name)
			log.Print(len(myOldFieldType))
		}

		actions = resourceTypeHandleEnumTypeChanges(newFieldType, oldFieldType, actions, name)
	}

	oldNames := make([]string, len(oldValues))
	newNames := make([]string, len(newValues))

	for i := range oldValues {
		v := oldValues[i].(map[string]interface{})
		oldNames[i] = v["name"].(string)
	}

	for i := range newValues {
		v := newValues[i].(map[string]interface{})
		newNames[i] = v["name"].(string)
	}

	if checkAttributeOrder && !reflect.DeepEqual(oldNames, newNames[:len(oldNames)]) {
		log.Printf("[DEBUG] Field ordering: %s", newNames)

		actions = append(
			actions,
			platform.TypeChangeFieldDefinitionOrderAction{
				FieldNames: newNames,
			})
	}

	return actions, nil
}

func resourceTypeHandleEnumTypeChanges(newFieldType platform.FieldType, oldFieldType map[string]interface{}, actions []platform.TypeUpdateAction, name string) []platform.TypeUpdateAction {
	if enumType, ok := newFieldType.(platform.CustomFieldEnumType); ok {
		oldEnumV := oldFieldType["values"].(map[string]interface{})

		for i := range enumType.Values {
			if _, ok := oldEnumV[enumType.Values[i].Key]; !ok {
				// Key does not appear in old enum values, so we'll add it
				actions = append(
					actions,
					platform.TypeAddEnumValueAction{
						FieldName: name,
						Value:     enumType.Values[i],
					})
				continue
			}

			if oldEnumV[enumType.Values[i].Key].(string) != enumType.Values[i].Label {
				//label for this key is changed
				actions = append(
					actions,
					platform.TypeChangeEnumValueLabelAction{
						FieldName: name,
						Value:     enumType.Values[i],
					})
			}
		}

		// Action: changeEnumValueOrder
		// TODO: Change the order of EnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-fielddefinitions

	} else if enumType, ok := newFieldType.(platform.CustomFieldLocalizedEnumType); ok {
		oldEnumV := oldFieldType["localized_value"].([]interface{})
		oldEnumKeys := make(map[string]map[string]interface{}, len(oldEnumV))

		for i := range oldEnumV {
			v := oldEnumV[i].(map[string]interface{})
			oldEnumKeys[v["key"].(string)] = v
		}

		for i, enumValue := range enumType.Values {
			if _, ok := oldEnumKeys[enumValue.Key]; !ok {
				// Key does not appear in old enum values, so we'll add it
				actions = append(
					actions,
					platform.TypeAddLocalizedEnumValueAction{
						FieldName: name,
						Value:     enumType.Values[i],
					})
			}
		}

		// Action: changeLocalizedEnumValueOrder
		// TODO: Change the order of LocalizedEnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-localizedenumvalues
	}
	return actions
}

func resourceTypeDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Types().WithId(d.Id()).Delete().Version(version).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTypeGetFieldDefinitions(d *schema.ResourceData) ([]platform.FieldDefinition, error) {
	input := d.Get("field").([]interface{})
	result := make([]platform.FieldDefinition, len(input))

	for i := range input {
		raw := input[i].(map[string]interface{})
		fieldDef, err := resourceTypeGetFieldDefinition(raw)

		if err != nil {
			return nil, err
		}

		result[i] = *fieldDef
	}

	return result, nil
}

func resourceTypeGetFieldDefinition(input map[string]interface{}) (*platform.FieldDefinition, error) {
	fieldTypes := input["type"].([]interface{})
	fieldType, err := getFieldType(fieldTypes[0])
	if err != nil {
		return nil, err
	}

	label := unmarshallLocalizedString(input["label"])
	inputHint := platform.TypeTextInputHint(input["input_hint"].(string))
	return &platform.FieldDefinition{
		Type:      fieldType,
		Name:      input["name"].(string),
		Label:     label,
		Required:  input["required"].(bool),
		InputHint: &inputHint,
	}, nil
}

func getFieldType(input interface{}) (platform.FieldType, error) {
	config := input.(map[string]interface{})
	typeName, ok := config["name"].(string)

	if !ok {
		return nil, fmt.Errorf("no 'name' for type object given")
	}

	switch typeName {
	case "Boolean":
		return platform.CustomFieldBooleanType{}, nil
	case "String":
		return platform.CustomFieldStringType{}, nil
	case "LocalizedString":
		return platform.CustomFieldLocalizedStringType{}, nil
	case "Enum":
		valuesInput, valuesOk := config["values"].(map[string]interface{})
		if !valuesOk {
			return nil, fmt.Errorf("no values specified for Enum type: %+v", valuesInput)
		}
		var values []platform.CustomFieldEnumValue
		for k, v := range valuesInput {
			values = append(values, platform.CustomFieldEnumValue{
				Key:   k,
				Label: v.(string),
			})
		}
		return platform.CustomFieldEnumType{Values: values}, nil
	case "LocalizedEnum":
		valuesInput, valuesOk := config["localized_value"]
		if !valuesOk {
			return nil, fmt.Errorf("no localized_value elements specified for LocalizedEnum type")
		}
		var values []platform.CustomFieldLocalizedEnumValue
		for _, value := range valuesInput.([]interface{}) {
			v := value.(map[string]interface{})
			labels := unmarshallLocalizedString(v["label"])
			values = append(values, platform.CustomFieldLocalizedEnumValue{
				Key:   v["key"].(string),
				Label: labels,
			})
		}
		return platform.CustomFieldLocalizedEnumType{Values: values}, nil
	case "Number":
		return platform.CustomFieldNumberType{}, nil
	case "Money":
		return platform.CustomFieldMoneyType{}, nil
	case "Date":
		return platform.CustomFieldDateType{}, nil
	case "Time":
		return platform.CustomFieldTimeType{}, nil
	case "DateTime":
		return platform.CustomFieldDateTimeType{}, nil
	case "Reference":
		refTypeID, refTypeIDOk := config["reference_type_id"].(string)
		if !refTypeIDOk {
			return nil, fmt.Errorf("no reference_type_id specified for Reference type")
		}
		return platform.CustomFieldReferenceType{
			ReferenceTypeId: platform.CustomFieldReferenceValue(refTypeID),
		}, nil
	case "Set":
		elementTypes, elementTypesOk := config["element_type"]
		if !elementTypesOk {
			return nil, fmt.Errorf("no element_type specified for Set type")
		}
		elementTypeList := elementTypes.([]interface{})
		if len(elementTypeList) == 0 {
			return nil, fmt.Errorf("no element_type specified for Set type")
		}

		setFieldType, err := getFieldType(elementTypeList[0])
		if err != nil {
			return nil, err
		}

		return platform.CustomFieldSetType{
			ElementType: setFieldType,
		}, nil
	}

	return nil, fmt.Errorf("unknown FieldType %s", typeName)
}

func marshallTypeFields(t *platform.Type) ([]map[string]interface{}, error) {
	fields := make([]map[string]interface{}, len(t.FieldDefinitions))
	for i, fieldDef := range t.FieldDefinitions {
		fieldData := make(map[string]interface{})
		log.Printf("[DEBUG] reading field: %s: %#v", fieldDef.Name, fieldDef)
		fieldType, err := marshallTypeFieldType(fieldDef.Type, true)
		if err != nil {
			return nil, err
		}
		fieldData["type"] = fieldType
		fieldData["name"] = fieldDef.Name
		if fieldDef.Label != nil {
			fieldData["label"] = fieldDef.Label
		} else {
			fieldData["label"] = nil
		}
		fieldData["required"] = fieldDef.Required
		fieldData["input_hint"] = fieldDef.InputHint

		fields[i] = fieldData
	}
	return fields, nil
}

func marshallTypeFieldType(fieldType platform.FieldType, setsAllowed bool) ([]interface{}, error) {
	typeData := make(map[string]interface{})

	switch val := fieldType.(type) {

	case platform.CustomFieldBooleanType:
		typeData["name"] = "Boolean"

	case platform.CustomFieldStringType:
		typeData["name"] = "String"

	case platform.CustomFieldLocalizedStringType:
		typeData["name"] = "LocalizedString"

	case platform.CustomFieldEnumType:
		enumValues := make(map[string]interface{}, len(val.Values))
		for _, value := range val.Values {
			enumValues[value.Key] = value.Label
		}
		typeData["name"] = "Enum"
		typeData["values"] = enumValues

	case platform.CustomFieldLocalizedEnumType:
		typeData["name"] = "LocalizedEnum"
		typeData["localized_value"] = marshallTypeLocalizedEnum(val.Values)

	case platform.CustomFieldNumberType:
		typeData["name"] = "Number"

	case platform.CustomFieldMoneyType:
		typeData["name"] = "Money"

	case platform.CustomFieldDateType:
		typeData["name"] = "Date"

	case platform.CustomFieldTimeType:
		typeData["name"] = "Time"

	case platform.CustomFieldDateTimeType:
		typeData["name"] = "DateTime"

	case platform.CustomFieldReferenceType:
		typeData["name"] = "Reference"
		typeData["reference_type_id"] = val.ReferenceTypeId

	case platform.CustomFieldSetType:
		typeData["name"] = "Set"
		if setsAllowed {
			elemType, err := marshallTypeFieldType(val.ElementType, false)
			if err != nil {
				return nil, err
			}
			typeData["element_type"] = elemType
		}

	default:
		return nil, fmt.Errorf("unknown resource Type %T: %#v", fieldType, fieldType)
	}

	return []interface{}{typeData}, nil
}

func marshallTypeLocalizedEnum(values []platform.CustomFieldLocalizedEnumValue) []interface{} {
	enumValues := make([]interface{}, len(values))
	for i := range values {
		enumValues[i] = map[string]interface{}{
			"key":   values[i].Key,
			"label": values[i].Label,
		}
	}
	return enumValues
}
