package commercetools

import (
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/hashicorp/terraform/helper/customdiff"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/service/types"
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
							Default:  commercetools.SingleLineTextInputHint,
						},
					},
				},
			},
			"version": &schema.Schema{
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
	svc := getTypeService(m)
	var ctType *types.Type

	name := commercetools.LocalizedString(
		expandStringMap(d.Get("name").(map[string]interface{})))
	description := commercetools.LocalizedString(
		expandStringMap(d.Get("description").(map[string]interface{})))

	resourceTypeIds := expandStringArray(
		d.Get("resource_type_ids").([]interface{}))
	fields, err := resourceTypeGetFieldDefinitions(d)

	if err != nil {
		return err
	}

	draft := &types.TypeDraft{
		Key:              d.Get("key").(string),
		Name:             name,
		Description:      description,
		ResourceTypeIds:  resourceTypeIds,
		FieldDefinitions: fields,
	}

	err = resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		ctType, err = svc.Create(draft)
		if err != nil {
			if ctErr, ok := err.(commercetools.Error); ok {
				if ctErr.Code() == commercetools.ErrInvalidJSONInput {
					return resource.NonRetryableError(ctErr)
				}
			} else {
				log.Printf("[DEBUG] Received error: %s", err)
			}
			return resource.RetryableError(err)
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
	svc := getTypeService(m)

	ctType, err := svc.GetByID(d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.Error); ok {
			if ctErr.Code() == commercetools.ErrResourceNotFound {
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
		d.Set("description", ctType.Description)
		d.Set("resource_type_ids", ctType.ResourceTypeIds)
		d.Set("field", fields)
	}
	return nil
}

func resourceTypeReadFieldType(fieldType types.FieldType, setsAllowed bool) ([]interface{}, error) {
	typeData := make(map[string]interface{})

	if _, ok := fieldType.(types.BooleanType); ok {
		typeData["name"] = "Boolean"
	} else if _, ok := fieldType.(types.StringType); ok {
		typeData["name"] = "String"
	} else if _, ok := fieldType.(types.LocalizedStringType); ok {
		typeData["name"] = "LocalizedString"
	} else if f, ok := fieldType.(types.EnumType); ok {
		enumValues := make(map[string]interface{}, len(f.Values))
		for _, value := range f.Values {
			enumValues[value.Key] = value.Label
		}
		typeData["name"] = "Enum"
		typeData["values"] = enumValues
	} else if f, ok := fieldType.(types.LocalizedEnumType); ok {
		typeData["name"] = "LocalizedEnum"
		typeData["localized_value"] = readLocalizedEnum(f.Values)
	} else if _, ok := fieldType.(types.NumberType); ok {
		typeData["name"] = "Number"
	} else if _, ok := fieldType.(types.MoneyType); ok {
		typeData["name"] = "Money"
	} else if _, ok := fieldType.(types.DateType); ok {
		typeData["name"] = "Date"
	} else if _, ok := fieldType.(types.TimeType); ok {
		typeData["name"] = "Time"
	} else if _, ok := fieldType.(types.DateTimeType); ok {
		typeData["name"] = "DateTime"
	} else if f, ok := fieldType.(types.ReferenceType); ok {
		typeData["name"] = "Reference"
		typeData["reference_type_id"] = f.ReferenceTypeID
	} else if f, ok := fieldType.(types.SetType); ok {
		typeData["name"] = "Set"
		if setsAllowed {
			elemType, err := resourceTypeReadFieldType(f.ElementType, false)
			if err != nil {
				return nil, err
			}
			typeData["element_type"] = elemType
		}
	} else {
		return nil, fmt.Errorf("Unkown resource Type %T", fieldType)
	}

	return []interface{}{typeData}, nil
}

func resourceTypeUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getTypeService(m)

	input := &types.UpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&types.ChangeKey{Key: newKey})
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(
			expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&types.ChangeName{Name: newName})
	}

	if d.HasChange("description") {
		newDescr := commercetools.LocalizedString(
			expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&types.SetDescription{Description: newDescr})
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

	_, err := svc.Update(input)
	if err != nil {
		if ctErr, ok := err.(commercetools.Error); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceTypeRead(d, m)
}

func resourceTypeFieldChangeActions(oldValues []interface{}, newValues []interface{}) ([]commercetools.UpdateAction, error) {
	oldLookup := createLookup(oldValues, "name")
	newLookup := createLookup(newValues, "name")
	actions := []commercetools.UpdateAction{}

	log.Printf("[DEBUG] Construction Field change actions")

	for name := range oldLookup {
		if _, ok := newLookup[name]; !ok {
			log.Printf("[DEBUG] Field deleted: %s", name)
			actions = append(actions, types.RemoveFieldDefinition{FieldName: name})
		}
	}

	for name, value := range newLookup {
		oldValue, existingField := oldLookup[name]
		newV := value.(map[string]interface{})

		fieldDef, err := resourceTypeGetFieldDefinition(newV)
		if err != nil {
			return nil, err
		}

		if !existingField {
			log.Printf("[DEBUG] Field added: %s", name)
			actions = append(
				actions,
				types.AddFieldDefinition{FieldDefinition: *fieldDef})
			continue
		}

		oldV := oldValue.(map[string]interface{})
		if !reflect.DeepEqual(oldV["label"], newV["label"]) {
			newLabel := commercetools.LocalizedString(
				expandStringMap(newV["label"].(map[string]interface{})))
			actions = append(
				actions,
				types.ChangeLabel{FieldName: name, Label: newLabel})
		}

		newFieldType := fieldDef.Type
		oldFieldType := oldV["type"].([]interface{})[0].(map[string]interface{})

		if enumType, ok := newFieldType.(types.EnumType); ok {
			oldEnumV := oldFieldType["values"].(map[string]interface{})

			for _, enumValue := range enumType.Values {
				if _, ok := oldEnumV[enumValue.Key]; !ok {
					// Key does not appear in old enum values, so we'll add it
					actions = append(
						actions,
						types.AddEnumValue{
							FieldName: name,
							Value:     enumValue,
						})
				}
			}

			// Action: changeEnumValueOrder
			// TODO: Change the order of EnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-fielddefinitions

		} else if enumType, ok := newFieldType.(types.LocalizedEnumType); ok {
			oldEnumV := oldFieldType["localized_value"].([]interface{})
			oldEnumKeys := make(map[string]map[string]interface{}, len(oldEnumV))

			for _, value := range oldEnumV {
				v := value.(map[string]interface{})
				oldEnumKeys[v["key"].(string)] = v
			}

			for _, enumValue := range enumType.Values {
				if _, ok := oldEnumKeys[enumValue.Key]; !ok {
					// Key does not appear in old enum values, so we'll add it
					actions = append(
						actions,
						types.AddLocalizedEnumValue{
							FieldName: name,
							Value:     enumValue,
						})
				}
			}

			// Action: changeLocalizedEnumValueOrder
			// TODO: Change the order of LocalizedEnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-localizedenumvalues
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

	if !reflect.DeepEqual(oldNames, newNames) {
		actions = append(
			actions,
			types.ChangeFieldDefinitionsOrder{
				FieldNames: newNames,
			})
	}

	return actions, nil
}

func resourceTypeDelete(d *schema.ResourceData, m interface{}) error {
	svc := getTypeService(m)
	version := d.Get("version").(int)
	_, err := svc.DeleteByID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func getTypeService(m interface{}) *types.Service {
	client := m.(*commercetools.Client)
	svc := types.New(client)
	return svc
}

func resourceTypeGetFieldDefinitions(d *schema.ResourceData) ([]types.FieldDefinition, error) {
	input := d.Get("field").([]interface{})
	var result []types.FieldDefinition

	for _, raw := range input {
		fieldDef, err := resourceTypeGetFieldDefinition(raw.(map[string]interface{}))

		if err != nil {
			return nil, err
		}

		result = append(result, *fieldDef)
	}

	return result, nil
}

func resourceTypeGetFieldDefinition(input map[string]interface{}) (*types.FieldDefinition, error) {
	fieldTypes := input["type"].([]interface{})
	fieldType, err := getFieldType(fieldTypes[0])
	if err != nil {
		return nil, err
	}

	label := commercetools.LocalizedString(
		expandStringMap(input["label"].(map[string]interface{})))

	return &types.FieldDefinition{
		Type:      fieldType,
		Name:      input["name"].(string),
		Label:     label,
		Required:  input["required"].(bool),
		InputHint: commercetools.TextInputHint(input["input_hint"].(string)),
	}, nil
}

func getFieldType(input interface{}) (types.FieldType, error) {
	config := input.(map[string]interface{})
	typeName, ok := config["name"].(string)

	if !ok {
		return nil, fmt.Errorf("No 'name' for type object given")
	}

	switch typeName {
	case "Boolean":
		return types.BooleanType{}, nil
	case "String":
		return types.StringType{}, nil
	case "LocalizedString":
		return types.LocalizedStringType{}, nil
	case "Enum":
		valuesInput, valuesOk := config["values"].(map[string]interface{})
		if !valuesOk {
			return nil, fmt.Errorf("No values specified for Enum type: %+v", valuesInput)
		}
		var values []commercetools.EnumValue
		for k, v := range valuesInput {
			values = append(values, commercetools.EnumValue{
				Key:   k,
				Label: v.(string),
			})
		}
		return types.EnumType{Values: values}, nil
	case "LocalizedEnum":
		valuesInput, valuesOk := config["localized_value"]
		if !valuesOk {
			return nil, fmt.Errorf("No localized_value elements specified for LocalizedEnum type")
		}
		var values []commercetools.LocalizedEnumValue
		for _, value := range valuesInput.([]interface{}) {
			v := value.(map[string]interface{})
			labels := expandStringMap(
				v["label"].(map[string]interface{}))
			values = append(values, commercetools.LocalizedEnumValue{
				Key:   v["key"].(string),
				Label: commercetools.LocalizedString(labels),
			})
		}
		return types.LocalizedEnumType{Values: values}, nil
	case "Number":
		return types.NumberType{}, nil
	case "Money":
		return types.MoneyType{}, nil
	case "Date":
		return types.DateType{}, nil
	case "Time":
		return types.TimeType{}, nil
	case "DateTime":
		return types.DateTimeType{}, nil
	case "Reference":
		refTypeID, refTypeIDOk := config["reference_type_id"].(string)
		if !refTypeIDOk {
			return nil, fmt.Errorf("No reference_type_id specified for Reference type")
		}
		return types.ReferenceType{
			ReferenceTypeID: refTypeID,
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

		return types.SetType{
			ElementType: setFieldType,
		}, nil
	}

	return nil, fmt.Errorf("Unkown FieldType %s", typeName)
}
