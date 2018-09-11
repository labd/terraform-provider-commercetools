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
				Type:     schema.TypeMap,
				Required: true,
			},
			"description": {
				Type:     schema.TypeMap,
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
							Type:     schema.TypeSet,
							Required: true,
							Elem:     fieldTypeElement(true),
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"label": {
							Type:     schema.TypeMap,
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
				oldLookup := resourceTypeCreateFieldLookup(old.([]interface{}))
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
					oldType := oldF["type"].(*schema.Set).List()[0].(map[string]interface{})
					newType := newF["type"].(*schema.Set).List()[0].(map[string]interface{})

					if oldType["name"] != newType["name"] {
						if oldType["name"] != "" || newType["name"] == "" {
							continue
						}
						return fmt.Errorf(
							"Field '%s' type changed from %s to %s. Changing types is not supported; please remove the field first and re-define it later",
							name, oldType["name"], newType["name"])
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
				Type:     schema.TypeMap,
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
			Type:     schema.TypeSet,
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
			Type:     schema.TypeSet,
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

func resourceTypeReadFieldType(fieldType types.FieldType, setsAllowed bool) (*schema.Set, error) {
	typeSchema := schema.HashResource(fieldTypeElement(setsAllowed))
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
		typeData["localized_value"] = resourceTypeReadLocalizedEnum(f.Values)
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
		elemType, err := resourceTypeReadFieldType(f.ElementType, false)
		if err != nil {
			return nil, err
		}
		if setsAllowed {
			typeData["element_type"] = elemType
		}
	} else {
		return nil, fmt.Errorf("Unkown resource Type %T", fieldType)
	}

	return schema.NewSet(typeSchema, []interface{}{typeData}), nil
}

func resourceTypeReadLocalizedEnum(values []commercetools.LocalizedEnumValue) *schema.Set {
	typeSchema := schema.HashResource(localizedValueElement())

	enumValues := make([]interface{}, len(values))
	for i, value := range values {
		enumValues[i] = map[string]interface{}{
			"key":   value.Key,
			"label": localizedStringToMap(value.Label),
		}
	}
	return schema.NewSet(typeSchema, enumValues)
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

func resourceTypeCreateFieldLookup(fields []interface{}) map[string]interface{} {
	lookup := make(map[string]interface{})
	for _, field := range fields {
		f := field.(map[string]interface{})
		lookup[f["name"].(string)] = field
	}
	return lookup
}

func resourceTypeFieldChangeActions(oldValues []interface{}, newValues []interface{}) ([]commercetools.UpdateAction, error) {
	oldLookup := resourceTypeCreateFieldLookup(oldValues)
	newLookup := resourceTypeCreateFieldLookup(newValues)
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
		oldV := oldValue.(map[string]interface{})

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

		if !reflect.DeepEqual(oldV["label"], newV["label"]) {
			newLabel := commercetools.LocalizedString(
				expandStringMap(newV["label"].(map[string]interface{})))
			actions = append(
				actions,
				types.ChangeLabel{FieldName: name, Label: newLabel})
		}

		// TODO: The following can result in some unexpected behaviour;
		// when a field type attribute changes, their hashes change and
		// terraform will concider this a resource being removed and a new one
		// being added.
		// Therefore it is quite tricky to compare changes.
		// This issue is further discussed here: https://github.com/hashicorp/terraform/issues/15420
		//
		// It might be better to define "type" as a `TypeList` instead of a `TypeSet`
		// with `MaxItems=1`.

		// newFieldType := fieldDef.Type
		// oldFieldType := oldV["type"].(*schema.Set).List()[0].(map[string]interface{})

		// if enumType, ok := newFieldType.(types.EnumType); ok {
		// 	oldEnumV := oldFieldType["values"].(map[string]interface{})

		// 	for _, enumValue := range enumType.Values {
		// 		if _, ok := oldEnumV[enumValue.Key]; !ok {
		// 			// Key does not appear in old enum values, so we'll add it
		// 			actions = append(
		// 				actions,
		// 				types.AddEnumValue{
		// 					FieldName: name,
		// 					Value:     enumValue,
		// 				})
		// 		}
		// 	}
		// } else if enumType, ok := newFieldType.(types.LocalizedEnumType); ok {
		// 	oldEnumV := oldFieldType["values"].(map[string]interface{})

		// 	for _, enumValue := range enumType.Values {
		// 		if _, ok := oldEnumV[enumValue.Key]; !ok {
		// 			// Key does not appear in old enum values, so we'll add it
		// 			actions = append(
		// 				actions,
		// 				types.AddLocalizedEnumValue{
		// 					FieldName: name,
		// 					Value:     enumValue,
		// 				})
		// 		}
		// 	}
		// }
	}

	// Action: changeFieldDefinitionOrder
	// TODO: Change the order of FieldDefinitions: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-fielddefinitions

	// Action: changeEnumValueOrder
	// TODO: Change the order of EnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-fielddefinitions

	// Action: changeLocalizedEnumValueOrder
	// TODO: Change the order of LocalizedEnumValues: https://docs.commercetools.com/http-api-projects-types.html#change-the-order-of-localizedenumvalues
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
	t, ok := input["type"].(*schema.Set)
	fieldTypes := []map[string]interface{}{}
	for _, ft := range t.List() {
		fieldType := ft.(map[string]interface{})
		// Field definitions gets reshuffled when one gets removed
		// or an attribute of the Field changes.
		// This results in having two types defined, one being empty (removed)
		// and one being an existing one (but moved to a 'new' field definition).
		if fieldType["name"] != "" {
			fieldTypes = append(fieldTypes, fieldType)
		}
	}

	if !ok {
		return nil, fmt.Errorf("No type defined for field definition")
	}
	if len(fieldTypes) > 1 {
		log.Printf("[DEBUG] %+v", fieldTypes)
		return nil, fmt.Errorf("More then 1 type definition detected. Please remove the redundant ones")
	}
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

func getFieldType(config map[string]interface{}) (types.FieldType, error) {
	typeName, ok := config["name"].(string)
	refTypeID, refTypeIDOk := config["reference_type_id"].(string)
	elementTypes, _ := config["element_type"].(*schema.Set)

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
		valuesInput, valuesOk := config["localized_value"].(*schema.Set)
		if !valuesOk {
			return nil, fmt.Errorf("No localized_value elements specified for LocalizedEnum type")
		}
		var values []commercetools.LocalizedEnumValue
		for _, value := range valuesInput.List() {
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
		if !refTypeIDOk {
			return nil, fmt.Errorf("No reference_type_id specified for Reference type")
		}
		return types.ReferenceType{
			ReferenceTypeID: refTypeID,
		}, nil
	case "Set":
		if elementTypes.Len() == 0 {
			return nil, fmt.Errorf("No element_type specified for Set type")
		} else if elementTypes.Len() > 1 {
			return nil, fmt.Errorf("Too many occurences of element_type for Set type. Only need 1")
		}

		setFieldType, err := getFieldType(elementTypes.List()[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		return types.SetType{
			ElementType: setFieldType,
		}, nil
	}

	return nil, fmt.Errorf("Unkown FieldType %s", typeName)
}
