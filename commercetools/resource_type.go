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
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceType() *schema.Resource {
	return &schema.Resource{
		Description: "Types define custom fields that are used to enhance resources as you need. Use Types to model " +
			"your own CustomFields on resources, like Category and Customer.\n\n" +
			"In case you want to customize products, please use product types instead that serve a similar purpose, " +
			"but tailored to products.\n\n" +
			"See also the [Types Api Documentation](https://docs.commercetools.com/api/projects/types)",
		Create: resourceTypeCreate,
		Read:   resourceTypeRead,
		Update: resourceTypeUpdate,
		Delete: resourceTypeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "Identifier for the type (max. 256 characters)",
				Type:        schema.TypeString,
				Required:    true,
			},
			"name": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Required:    true,
			},
			"description": {
				Description: "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:        TypeLocalizedString,
				Optional:    true,
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
							Description: "A human-readable label for the field",
							Type:        TypeLocalizedString,
							Required:    true,
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
						if oldType["name"] != "" || newType["name"] == "" {
							continue
						}
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
	var ctType *platform.Type

	name := platform.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := platform.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	resourceTypeIds := []platform.ResourceTypeId{}
	for _, item := range expandStringArray(d.Get("resource_type_ids").([]interface{})) {
		resourceTypeIds = append(resourceTypeIds, platform.ResourceTypeId(item))

	}
	fields, err := resourceTypeGetFieldDefinitions(d)

	if err != nil {
		return err
	}

	draft := platform.TypeDraft{
		Key:              d.Get("key").(string),
		Name:             name,
		Description:      &description,
		ResourceTypeIds:  resourceTypeIds,
		FieldDefinitions: fields,
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		ctType, err = client.Types().Post(draft).Execute(context.Background())
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

	ctType, err := client.Types().WithId(d.Id()).Get().Execute(context.Background())

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
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
			fieldData["label"] = fieldDef.Label
			fieldData["required"] = fieldDef.Required
			fieldData["input_hint"] = fieldDef.InputHint

			fields[i] = fieldData
		}

		d.Set("version", ctType.Version)
		d.Set("key", ctType.Key)
		d.Set("name", ctType.Name)
		if ctType.Description != nil {
			d.Set("description", ctType.Description)
		}
		d.Set("resource_type_ids", ctType.ResourceTypeIds)
		d.Set("field", fields)
	}
	return nil
}

func resourceTypeReadFieldType(fieldType platform.FieldType, setsAllowed bool) ([]interface{}, error) {
	typeData := make(map[string]interface{})

	if _, ok := fieldType.(platform.CustomFieldBooleanType); ok {
		typeData["name"] = "Boolean"
	} else if _, ok := fieldType.(platform.CustomFieldStringType); ok {
		typeData["name"] = "String"
	} else if _, ok := fieldType.(platform.CustomFieldLocalizedStringType); ok {
		typeData["name"] = "LocalizedString"
	} else if f, ok := fieldType.(platform.CustomFieldEnumType); ok {
		enumValues := make(map[string]interface{}, len(f.Values))
		for _, value := range f.Values {
			enumValues[value.Key] = value.Label
		}
		typeData["name"] = "Enum"
		typeData["values"] = enumValues
	} else if f, ok := fieldType.(platform.CustomFieldLocalizedEnumType); ok {
		typeData["name"] = "LocalizedEnum"
		typeData["localized_value"] = readCustomFieldLocalizedEnum(f.Values)
	} else if _, ok := fieldType.(platform.CustomFieldNumberType); ok {
		typeData["name"] = "Number"
	} else if _, ok := fieldType.(platform.CustomFieldMoneyType); ok {
		typeData["name"] = "Money"
	} else if _, ok := fieldType.(platform.CustomFieldDateType); ok {
		typeData["name"] = "Date"
	} else if _, ok := fieldType.(platform.CustomFieldTimeType); ok {
		typeData["name"] = "Time"
	} else if _, ok := fieldType.(platform.CustomFieldDateTimeType); ok {
		typeData["name"] = "DateTime"
	} else if f, ok := fieldType.(platform.CustomFieldReferenceType); ok {
		typeData["name"] = "Reference"
		typeData["reference_type_id"] = f.ReferenceTypeId
	} else if f, ok := fieldType.(platform.CustomFieldSetType); ok {
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
		newName := platform.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&platform.TypeChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescr := platform.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&platform.TypeSetDescriptionAction{
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

	_, err := client.Types().WithId(d.Id()).Post(input).Execute(context.Background())
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceTypeRead(d, m)
}

// Generate a list of actions needed for updating the fields value in
// commercetools so that it matches the terraform file
func resourceTypeFieldChangeActions(oldValues []interface{}, newValues []interface{}) ([]platform.TypeUpdateAction, error) {
	oldLookup := createLookup(oldValues, "name")
	newLookup := createLookup(newValues, "name")
	actions := []platform.TypeUpdateAction{}
	checkAttributeOrder := true

	log.Printf("[DEBUG] Construction Field change actions")

	// Check if we have fields which are removed
	for name := range oldLookup {
		if _, ok := newLookup[name]; !ok {
			log.Printf("[DEBUG] Field deleted: %s", name)
			actions = append(actions, platform.TypeRemoveFieldDefinitionAction{FieldName: name})
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
				platform.TypeAddFieldDefinitionAction{FieldDefinition: *fieldDef})
			checkAttributeOrder = false
			continue
		}

		// Check if we need to update the field label
		oldV := oldValue.(map[string]interface{})
		if !reflect.DeepEqual(oldV["label"], newV["label"]) {
			newLabel := platform.LocalizedString(
				expandStringMap(newV["label"].(map[string]interface{})))
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
			platform.TypeChangeFieldDefinitionOrderAction{
				FieldNames: newNames,
			})
	}

	return actions, nil
}

func resourceTypeHandleEnumTypeChanges(newFieldType platform.FieldType, oldFieldType map[string]interface{}, actions []platform.TypeUpdateAction, name string) []platform.TypeUpdateAction {
	if enumType, ok := newFieldType.(platform.CustomFieldEnumType); ok {
		oldEnumV := oldFieldType["values"].(map[string]interface{})

		for i, enumValue := range enumType.Values {
			if _, ok := oldEnumV[enumValue.Key]; !ok {
				// Key does not appear in old enum values, so we'll add it
				actions = append(
					actions,
					platform.TypeAddEnumValueAction{
						FieldName: name,
						Value:     enumType.Values[i],
					})
				continue
			}

			if oldEnumV[enumValue.Key].(string) != enumValue.Label {
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

		for _, value := range oldEnumV {
			v := value.(map[string]interface{})
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

func resourceTypeDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	version := d.Get("version").(int)
	_, err := client.Types().WithId(d.Id()).Delete().WithQueryParams(platform.ByProjectKeyTypesByIDRequestMethodDeleteInput{
		Version: version,
	}).Execute(context.Background())
	if err != nil {
		return err
	}

	return nil
}

func resourceTypeGetFieldDefinitions(d *schema.ResourceData) ([]platform.FieldDefinition, error) {
	input := d.Get("field").([]interface{})
	var result []platform.FieldDefinition

	for _, raw := range input {
		fieldDef, err := resourceTypeGetFieldDefinition(raw.(map[string]interface{}))

		if err != nil {
			return nil, err
		}

		result = append(result, *fieldDef)
	}

	return result, nil
}

func resourceTypeGetFieldDefinition(input map[string]interface{}) (*platform.FieldDefinition, error) {
	fieldTypes := input["type"].([]interface{})
	fieldType, err := getFieldType(fieldTypes[0])
	if err != nil {
		return nil, err
	}

	label := platform.LocalizedString(
		expandStringMap(input["label"].(map[string]interface{})))

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
		return nil, fmt.Errorf("No 'name' for type object given")
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
			return nil, fmt.Errorf("No values specified for Enum type: %+v", valuesInput)
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
			return nil, fmt.Errorf("No localized_value elements specified for LocalizedEnum type")
		}
		var values []platform.CustomFieldLocalizedEnumValue
		for _, value := range valuesInput.([]interface{}) {
			v := value.(map[string]interface{})
			labels := platform.LocalizedString(
				expandStringMap(v["label"].(map[string]interface{})))
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
			return nil, fmt.Errorf("No reference_type_id specified for Reference type")
		}
		return platform.CustomFieldReferenceType{
			ReferenceTypeId: platform.ReferenceTypeId(refTypeID),
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

		return platform.CustomFieldSetType{
			ElementType: setFieldType,
		}, nil
	}

	return nil, fmt.Errorf("Unknown FieldType %s", typeName)
}

func readCustomFieldLocalizedEnum(values []platform.CustomFieldLocalizedEnumValue) []interface{} {
	enumValues := make([]interface{}, len(values))
	for i, value := range values {
		enumValues[i] = map[string]interface{}{
			"key":   value.Key,
			"label": &value.Label,
		}
	}
	return enumValues
}
