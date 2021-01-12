package commercetools

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceType() *schema.Resource {
	return &schema.Resource{
		Create: resourceTypeCreate,
		Read:   resourceTypeRead,
		Update: resourceTypeUpdate,
		Delete: resourceTypeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     TypeLocalizedString,
				Required: true,
			},
			"description": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"resource_type_ids": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"field": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem:     fieldTypeElement(true),
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"label": {
							Type:     TypeLocalizedString,
							Required: true,
						},
						"required": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"input_hint": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  commercetools.TextInputHintSingleLine,
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
			customdiff.ValidateChange("field", func(old, new, meta interface{}) error {
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
						return fmt.Errorf(
							"Field '%s' type changed from %s to %s. Changing types is not supported; please remove the field first and re-define it later",
							name, oldType["name"], newType["name"])
					}

					if oldF["required"] != newF["required"] {
						return fmt.Errorf(
							"Error on the '%s' attribute: Updating the 'required' attribute is not supported. Consider removing the attribute first and then re-adding it",
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
				Type:     TypeLocalizedString,
				Required: true,
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
					errs = append(errs, fmt.Errorf("Sets in another Set are not allowed"))
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

func resourceTypeCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var ctType *commercetools.Type

	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := commercetools.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	resourceTypeIds := []commercetools.ResourceTypeID{}
	for _, item := range expandStringArray(d.Get("resource_type_ids").([]interface{})) {
		resourceTypeIds = append(resourceTypeIds, commercetools.ResourceTypeID(item))

	}
	fields, err := resourceTypeGetFieldDefinitions(d)

	if err != nil {
		return err
	}

	draft := &commercetools.TypeDraft{
		Key:              d.Get("key").(string),
		Name:             &name,
		Description:      &description,
		ResourceTypeIds:  resourceTypeIds,
		FieldDefinitions: fields,
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		ctType, err = client.TypeCreate(context.Background(), draft)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if ctType == nil {
		log.Fatal("No type created?")
	}

	d.SetId(ctType.ID)
	d.Set("version", ctType.Version)

	return resourceTypeRead(d, m)
}

func resourceTypeRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading type from commercetools")
	client := getClient(m)

	ctType, err := client.TypeGetWithID(context.Background(), d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if ctType == nil {
		log.Print("[DEBUG] No type found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following type:")
		log.Print(stringFormatObject(ctType))

		fields := make([]map[string]interface{}, len(ctType.FieldDefinitions))
		for i, fieldDef := range ctType.FieldDefinitions {
			fieldData := make(map[string]interface{})
			log.Printf("[DEBUG] reading field: %s: %#v", fieldDef.Name, fieldDef)
			fieldType, err := resourceTypeReadFieldType(fieldDef.Type, true)
			if err != nil {
				return err
			}
			fieldData["type"] = fieldType
			fieldData["name"] = fieldDef.Name
			fieldData["label"] = *fieldDef.Label
			fieldData["required"] = fieldDef.Required
			fieldData["input_hint"] = fieldDef.InputHint

			fields[i] = fieldData
		}

		d.Set("version", ctType.Version)
		d.Set("key", ctType.Key)
		d.Set("name", *ctType.Name)
		if ctType.Description != nil {
			d.Set("description", ctType.Description)
		}
		d.Set("resource_type_ids", ctType.ResourceTypeIds)
		d.Set("field", fields)
	}
	return nil
}

func resourceTypeReadFieldType(fieldType commercetools.FieldType, setsAllowed bool) ([]interface{}, error) {
	typeData := make(map[string]interface{})

	if _, ok := fieldType.(commercetools.CustomFieldBooleanType); ok {
		typeData["name"] = "Boolean"
	} else if _, ok := fieldType.(commercetools.CustomFieldStringType); ok {
		typeData["name"] = "String"
	} else if _, ok := fieldType.(commercetools.CustomFieldLocalizedStringType); ok {
		typeData["name"] = "LocalizedString"
	} else if f, ok := fieldType.(commercetools.CustomFieldEnumType); ok {
		enumValues := make(map[string]interface{}, len(f.Values))
		for _, value := range f.Values {
			enumValues[value.Key] = value.Label
		}
		typeData["name"] = "Enum"
		typeData["values"] = enumValues
	} else if f, ok := fieldType.(commercetools.CustomFieldLocalizedEnumType); ok {
		typeData["name"] = "LocalizedEnum"
		typeData["localized_value"] = readCustomFieldLocalizedEnum(f.Values)
	} else if _, ok := fieldType.(commercetools.CustomFieldNumberType); ok {
		typeData["name"] = "Number"
	} else if _, ok := fieldType.(commercetools.CustomFieldMoneyType); ok {
		typeData["name"] = "Money"
	} else if _, ok := fieldType.(commercetools.CustomFieldDateType); ok {
		typeData["name"] = "Date"
	} else if _, ok := fieldType.(commercetools.CustomFieldTimeType); ok {
		typeData["name"] = "Time"
	} else if _, ok := fieldType.(commercetools.CustomFieldDateTimeType); ok {
		typeData["name"] = "DateTime"
	} else if f, ok := fieldType.(commercetools.CustomFieldReferenceType); ok {
		typeData["name"] = "Reference"
		typeData["reference_type_id"] = f.ReferenceTypeID
	} else if f, ok := fieldType.(commercetools.CustomFieldSetType); ok {
		typeData["name"] = "Set"
		if setsAllowed {
			elemType, err := resourceTypeReadFieldType(f.ElementType, false)
			if err != nil {
				return nil, err
			}
			typeData["element_type"] = elemType
		}
	} else {
		return nil, fmt.Errorf("Unknown resource Type %T: %#v", fieldType, fieldType)
	}

	return []interface{}{typeData}, nil
}

func resourceTypeUpdate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	input := &commercetools.TypeUpdateWithIDInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: []commercetools.TypeUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.TypeChangeKeyAction{Key: newKey})
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.TypeChangeNameAction{Name: &newName})
	}

	if d.HasChange("description") {
		newDescr := commercetools.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.TypeSetDescriptionAction{
				Description: &newDescr})
	}

	if d.HasChange("field") {
		old, new := d.GetChange("field")
		fieldChangeActions, err := resourceTypeFieldChangeActions(old.([]interface{}), new.([]interface{}))
		if err != nil {
			return err
		}
		input.Actions = append(input.Actions, fieldChangeActions...)
	}
	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err := client.TypeUpdateWithID(context.Background(), input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceTypeRead(d, m)
}

// Generate a list of actions needed for updating the fields value in
// commercetools so that it matches the terraform file
func resourceTypeFieldChangeActions(oldValues []interface{}, newValues []interface{}) ([]commercetools.TypeUpdateAction, error) {
	oldLookup := createLookup(oldValues, "name")
	newLookup := createLookup(newValues, "name")
	actions := []commercetools.TypeUpdateAction{}
	checkAttributeOrder := true

	log.Printf("[DEBUG] Construction Field change actions")

	// Check if we have fields which are removed
	for name := range oldLookup {
		if _, ok := newLookup[name]; !ok {
			log.Printf("[DEBUG] Field deleted: %s", name)
			actions = append(actions, commercetools.TypeRemoveFieldDefinitionAction{FieldName: name})
			checkAttributeOrder = false
		}
	}

	for _, value := range newValues {
		newV := value.(map[string]interface{})
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
				commercetools.TypeAddFieldDefinitionAction{FieldDefinition: fieldDef})
			checkAttributeOrder = false
			continue
		}

		// Check if we need to update the field label
		oldV := oldValue.(map[string]interface{})
		if !reflect.DeepEqual(oldV["label"], newV["label"]) {
			newLabel := commercetools.LocalizedString(
				expandStringMap(newV["label"].(map[string]interface{})))
			actions = append(
				actions,
				commercetools.TypeChangeLabelAction{FieldName: name, Label: &newLabel})
		}

		// Update the input hint if this is changed
		if !reflect.DeepEqual(oldV["input_hint"], newV["input_hint"]) {
			var newInputHint commercetools.TypeTextInputHint
			switch newV["input_hint"].(string) {
			case "SingleLine":
				newInputHint = commercetools.TypeTextInputHintSingleLine
			case "MultiLine":
				newInputHint = commercetools.TypeTextInputHintMultiLine
			}

			actions = append(
				actions,
				commercetools.TypeChangeInputHintAction{FieldName: name, InputHint: newInputHint})
		}

		newFieldType := fieldDef.Type
		oldFieldType := oldV["type"].([]interface{})[0].(map[string]interface{})

		if enumType, ok := newFieldType.(commercetools.CustomFieldSetType); ok {

			myOldFieldType := oldFieldType["element_type"].([]interface{})[0].(map[string]interface{})
			actions = resourceTypeHandleEnumTypeChanges(enumType.ElementType, myOldFieldType, actions, name)

			log.Printf("[DEBUG] Set detected: %s", name)
			log.Print(len(myOldFieldType))
		}

		actions = resourceTypeHandleEnumTypeChanges(newFieldType, oldFieldType, actions, name)
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
			commercetools.TypeChangeFieldDefinitionOrderAction{
				FieldNames: newNames,
			})
	}

	return actions, nil
}

func resourceTypeHandleEnumTypeChanges(newFieldType commercetools.FieldType, oldFieldType map[string]interface{}, actions []commercetools.TypeUpdateAction, name string) []commercetools.TypeUpdateAction {
	if enumType, ok := newFieldType.(commercetools.CustomFieldEnumType); ok {
		oldEnumV := oldFieldType["values"].(map[string]interface{})

		for i, enumValue := range enumType.Values {
			if _, ok := oldEnumV[enumValue.Key]; !ok {
				// Key does not appear in old enum values, so we'll add it
				actions = append(
					actions,
					commercetools.TypeAddEnumValueAction{
						FieldName: name,
						Value:     &enumType.Values[i],
					})
				continue
			}

			if oldEnumV[enumValue.Key].(string) != enumValue.Label {
				//label for this key is changed
				actions = append(
					actions,
					commercetools.TypeChangeEnumValueLabelAction{
						FieldName: name,
						Value:     &enumType.Values[i],
					})
			}
		}

		// Action: changeEnumValueOrder
		// TODO: Change the order of EnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-fielddefinitions

	} else if enumType, ok := newFieldType.(commercetools.CustomFieldLocalizedEnumType); ok {
		oldEnumV := oldFieldType["localized_value"].([]interface{})
		oldEnumKeys := make(map[string]map[string]interface{}, len(oldEnumV))

		for _, value := range oldEnumV {
			v := value.(map[string]interface{})
			oldEnumKeys[v["key"].(string)] = v
		}

		for i, enumValue := range enumType.Values {
			if _, ok := oldEnumKeys[enumValue.Key]; !ok {
				// Key does not appear in old enum values, so we'll add it
				actions = append(
					actions,
					commercetools.TypeAddLocalizedEnumValueAction{
						FieldName: name,
						Value:     &enumType.Values[i],
					})
			}
		}

		// Action: changeLocalizedEnumValueOrder
		// TODO: Change the order of LocalizedEnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-localizedenumvalues
	}
	return actions
}

func resourceTypeDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.TypeDeleteWithID(context.Background(), d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func resourceTypeGetFieldDefinitions(d *schema.ResourceData) ([]commercetools.FieldDefinition, error) {
	input := d.Get("field").([]interface{})
	var result []commercetools.FieldDefinition

	for _, raw := range input {
		fieldDef, err := resourceTypeGetFieldDefinition(raw.(map[string]interface{}))

		if err != nil {
			return nil, err
		}

		result = append(result, *fieldDef)
	}

	return result, nil
}

func resourceTypeGetFieldDefinition(input map[string]interface{}) (*commercetools.FieldDefinition, error) {
	fieldTypes := input["type"].([]interface{})
	fieldType, err := getFieldType(fieldTypes[0])
	if err != nil {
		return nil, err
	}

	label := commercetools.LocalizedString(
		expandStringMap(input["label"].(map[string]interface{})))

	return &commercetools.FieldDefinition{
		Type:      fieldType,
		Name:      input["name"].(string),
		Label:     &label,
		Required:  input["required"].(bool),
		InputHint: commercetools.TypeTextInputHint(input["input_hint"].(string)),
	}, nil
}

func getFieldType(input interface{}) (commercetools.FieldType, error) {
	config := input.(map[string]interface{})
	typeName, ok := config["name"].(string)

	if !ok {
		return nil, fmt.Errorf("No 'name' for type object given")
	}

	switch typeName {
	case "Boolean":
		return commercetools.CustomFieldBooleanType{}, nil
	case "String":
		return commercetools.CustomFieldStringType{}, nil
	case "LocalizedString":
		return commercetools.CustomFieldLocalizedStringType{}, nil
	case "Enum":
		valuesInput, valuesOk := config["values"].(map[string]interface{})
		if !valuesOk {
			return nil, fmt.Errorf("No values specified for Enum type: %+v", valuesInput)
		}
		var values []commercetools.CustomFieldEnumValue
		for k, v := range valuesInput {
			values = append(values, commercetools.CustomFieldEnumValue{
				Key:   k,
				Label: v.(string),
			})
		}
		return commercetools.CustomFieldEnumType{Values: values}, nil
	case "LocalizedEnum":
		valuesInput, valuesOk := config["localized_value"]
		if !valuesOk {
			return nil, fmt.Errorf("No localized_value elements specified for LocalizedEnum type")
		}
		var values []commercetools.CustomFieldLocalizedEnumValue
		for _, value := range valuesInput.([]interface{}) {
			v := value.(map[string]interface{})
			labels := commercetools.LocalizedString(
				expandStringMap(v["label"].(map[string]interface{})))
			values = append(values, commercetools.CustomFieldLocalizedEnumValue{
				Key:   v["key"].(string),
				Label: &labels,
			})
		}
		return commercetools.CustomFieldLocalizedEnumType{Values: values}, nil
	case "Number":
		return commercetools.CustomFieldNumberType{}, nil
	case "Money":
		return commercetools.CustomFieldMoneyType{}, nil
	case "Date":
		return commercetools.CustomFieldDateType{}, nil
	case "Time":
		return commercetools.CustomFieldTimeType{}, nil
	case "DateTime":
		return commercetools.CustomFieldDateTimeType{}, nil
	case "Reference":
		refTypeID, refTypeIDOk := config["reference_type_id"].(string)
		if !refTypeIDOk {
			return nil, fmt.Errorf("No reference_type_id specified for Reference type")
		}
		return commercetools.CustomFieldReferenceType{
			ReferenceTypeID: commercetools.ReferenceTypeID(refTypeID),
		}, nil
	case "Set":
		elementTypes, elementTypesOk := config["element_type"]
		if !elementTypesOk {
			return nil, fmt.Errorf("No element_type specified for Set type")
		}
		elementTypeList := elementTypes.([]interface{})
		if len(elementTypeList) == 0 {
			return nil, fmt.Errorf("No element_type specified for Set type")
		}

		setFieldType, err := getFieldType(elementTypeList[0])
		if err != nil {
			return nil, err
		}

		return commercetools.CustomFieldSetType{
			ElementType: setFieldType,
		}, nil
	}

	return nil, fmt.Errorf("Unknown FieldType %s", typeName)
}

func readCustomFieldLocalizedEnum(values []commercetools.CustomFieldLocalizedEnumValue) []interface{} {
	enumValues := make([]interface{}, len(values))
	for i, value := range values {
		enumValues[i] = map[string]interface{}{
			"key":   value.Key,
			"label": &value.Label,
		}
	}
	return enumValues
}
