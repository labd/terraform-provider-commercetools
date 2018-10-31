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
	"github.com/labd/commercetools-go-sdk/service/producttypes"
)

var constraintMap = map[string]producttypes.AttributeConstraint{
	"Unique":            producttypes.UniqueAttributeConstraint,
	"CombinationUnique": producttypes.CombinationUniqueAttributeConstraint,
	"SameForAll":        producttypes.SameForAllAttributeConstraint,
	"None":              producttypes.NoneAttributeConstraint,
}

func resourceProductType() *schema.Resource {
	return &schema.Resource{
		Create: resourceProductTypeCreate,
		Read:   resourceProductTypeRead,
		Update: resourceProductTypeUpdate,
		Delete: resourceProductTypeDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
				Type:     schema.TypeString,
				Optional: true,
			},
			"attribute": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeList,
							MaxItems: 1,
							Required: true,
							Elem:     attributeTypeElement(true),
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
						"constraint": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  producttypes.NoneAttributeConstraint,
							ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
								v := val.(string)

								if _, ok := constraintMap[v]; !ok {
									allowedConstraints := []string{}
									for key := range constraintMap {
										allowedConstraints = append(allowedConstraints, key)
									}
									errs = append(errs, fmt.Errorf(
										"Unkown attribute constraint '%v'. Possible values are %v", v, allowedConstraints))
								}
								return
							},
						},
						"input_tip": {
							Type:     TypeLocalizedString,
							Optional: true,
						},
						"input_hint": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  commercetools.SingleLineTextInputHint,
						},
						"searchable": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
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
			customdiff.ValidateChange("attribute", func(old, new, meta interface{}) error {
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
							"Field '%s' type changed from %s to %s. Changing types is not supported; please remove the attribute first and re-define it later",
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

func attributeTypeElement(setsAllowed bool) *schema.Resource {
	result := map[string]*schema.Schema{
		"name": {
			Type:     schema.TypeString,
			Required: true,
			ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
				v := val.(string)
				if !setsAllowed && v == "set" {
					errs = append(errs, fmt.Errorf("Sets in another Set are not allowed"))
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
			Elem:     attributeTypeElement(false),
		}
	}

	return &schema.Resource{Schema: result}
}

func resourceProductTypeCreate(d *schema.ResourceData, m interface{}) error {
	svc := getProductTypeService(m)
	var ctType *producttypes.ProductType

	attributes, err := resourceProductTypeGetAttributeDefinitions(d)

	if err != nil {
		return err
	}

	draft := &producttypes.ProductTypeDraft{
		Key:         d.Get("key").(string),
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Attributes:  attributes,
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

	return resourceProductTypeRead(d, m)
}

func resourceProductTypeRead(d *schema.ResourceData, m interface{}) error {
	log.Print("[DEBUG] Reading product type from commercetools")
	svc := getProductTypeService(m)

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
		log.Print("[DEBUG] No product type found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following product type:")
		log.Print(stringFormatObject(ctType))

		attributes := make([]map[string]interface{}, len(ctType.Attributes))
		for i, fieldDef := range ctType.Attributes {
			fieldData := make(map[string]interface{})
			fieldType, err := resourceProductTypeReadAttributeType(fieldDef.Type, true)
			if err != nil {
				return err
			}

			fieldData["type"] = fieldType
			fieldData["name"] = fieldDef.Name
			fieldData["label"] = fieldDef.Label
			fieldData["required"] = fieldDef.IsRequired
			fieldData["input_hint"] = fieldDef.InputHint
			fieldData["input_tip"] = fieldDef.InputTip
			fieldData["constraint"] = fieldDef.AttributeConstraint
			fieldData["searchable"] = fieldDef.IsSearchable

			attributes[i] = fieldData
		}

		d.Set("version", ctType.Version)
		d.Set("key", ctType.Key)
		d.Set("name", ctType.Name)
		d.Set("description", ctType.Description)
		d.Set("attribute", attributes)
	}
	return nil
}

func resourceProductTypeReadAttributeType(attrType producttypes.AttributeType, setsAllowed bool) ([]interface{}, error) {
	typeData := make(map[string]interface{})

	if _, ok := attrType.(producttypes.BooleanType); ok {
		typeData["name"] = "boolean"
	} else if _, ok := attrType.(producttypes.TextType); ok {
		typeData["name"] = "text"
	} else if _, ok := attrType.(producttypes.LocalizedTextType); ok {
		typeData["name"] = "ltext"
	} else if f, ok := attrType.(producttypes.EnumType); ok {
		enumValues := make(map[string]interface{}, len(f.Values))
		for _, value := range f.Values {
			enumValues[value.Key] = value.Label
		}
		typeData["name"] = "enum"
		typeData["values"] = enumValues
	} else if f, ok := attrType.(producttypes.LocalizedEnumType); ok {
		typeData["name"] = "lenum"
		typeData["localized_value"] = readLocalizedEnum(f.Values)
	} else if _, ok := attrType.(producttypes.NumberType); ok {
		typeData["name"] = "number"
	} else if _, ok := attrType.(producttypes.MoneyType); ok {
		typeData["name"] = "money"
	} else if _, ok := attrType.(producttypes.DateType); ok {
		typeData["name"] = "date"
	} else if _, ok := attrType.(producttypes.TimeType); ok {
		typeData["name"] = "time"
	} else if _, ok := attrType.(producttypes.DateTimeType); ok {
		typeData["name"] = "datetime"
	} else if f, ok := attrType.(producttypes.ReferenceType); ok {
		typeData["name"] = "reference"
		typeData["reference_type_id"] = f.ReferenceTypeID
	} else if f, ok := attrType.(producttypes.NestedType); ok {
		typeData["name"] = "nested"
		typeData["reference_type_id"] = f.TypeReference.ID
	} else if f, ok := attrType.(producttypes.SetType); ok {
		typeData["name"] = "set"
		if setsAllowed {
			elemType, err := resourceTypeReadFieldType(f.ElementType, false)
			if err != nil {
				return nil, err
			}
			typeData["element_type"] = elemType
		}
	} else {
		return nil, fmt.Errorf("Unkown resource Type %T", attrType)
	}

	return []interface{}{typeData}, nil
}

func resourceProductTypeUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getProductTypeService(m)

	input := &producttypes.UpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&producttypes.SetKey{Key: newKey})
	}

	if d.HasChange("name") {
		newName := d.Get("name").(string)
		input.Actions = append(
			input.Actions,
			&producttypes.ChangeName{Name: newName})
	}

	if d.HasChange("description") {
		newDescr := d.Get("description").(string)
		input.Actions = append(
			input.Actions,
			&producttypes.ChangeDescription{Description: newDescr})
	}

	if d.HasChange("attribute") {
		old, new := d.GetChange("attribute")
		attributeChangeActions, err := resourceProductTypeAttributeChangeActions(
			old.([]interface{}), new.([]interface{}))
		if err != nil {
			return err
		}

		input.Actions = append(input.Actions, attributeChangeActions...)
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

	return resourceProductTypeRead(d, m)
}

func resourceProductTypeDelete(d *schema.ResourceData, m interface{}) error {
	svc := getProductTypeService(m)
	version := d.Get("version").(int)
	_, err := svc.DeleteByID(d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func resourceProductTypeAttributeChangeActions(oldValues []interface{}, newValues []interface{}) ([]commercetools.UpdateAction, error) {
	oldLookup := createLookup(oldValues, "name")
	newLookup := createLookup(newValues, "name")
	newAttrDefinitions := []producttypes.AttributeDefinition{}
	actions := []commercetools.UpdateAction{}
	checkAttributeOrder := true

	for name := range oldLookup {
		if _, ok := newLookup[name]; !ok {
			log.Printf("[DEBUG] Attribute deleted: %s", name)
			actions = append(actions, producttypes.RemoveAttributeDefinition{Name: name})
			checkAttributeOrder = false
		}
	}

	for name, value := range newLookup {
		oldValue, existingField := oldLookup[name]
		newV := value.(map[string]interface{})

		var attrDef producttypes.AttributeDefinition
		if output, err := resourceProductTypeGetAttributeDefinition(newV, false); err == nil {
			attrDef = output.(producttypes.AttributeDefinition)
		} else {
			return nil, err
		}

		var attrDefDraft producttypes.AttributeDefinitionDraft
		if output, err := resourceProductTypeGetAttributeDefinition(newV, true); err == nil {
			attrDefDraft = output.(producttypes.AttributeDefinitionDraft)
		} else {
			return nil, err
		}

		newAttrDefinitions = append(newAttrDefinitions, attrDef)

		if !existingField {
			log.Printf("[DEBUG] Attribute added: %s", name)
			actions = append(
				actions,
				producttypes.AddAttributeDefinition{Attribute: attrDefDraft})
			checkAttributeOrder = false
			continue
		}

		oldV := oldValue.(map[string]interface{})
		if !reflect.DeepEqual(oldV["label"], newV["label"]) {
			actions = append(
				actions,
				producttypes.ChangeAttributeDefinitionLabel{
					AttributeName: name, Label: attrDef.Label})
		}
		if oldV["name"] != newV["name"] {
			actions = append(
				actions,
				producttypes.ChangeAttributeDefinitionName{
					AttributeName: name, NewAttributeName: attrDef.Name})
		}
		if oldV["searchable"] != newV["searchable"] {
			actions = append(
				actions,
				producttypes.ChangeIsSearchable{
					AttributeName: name, IsSearchable: attrDef.IsSearchable})
		}
		if oldV["input_hint"] != newV["input_hint"] {
			actions = append(
				actions,
				producttypes.ChangeInputHint{
					AttributeName: name, NewValue: attrDef.InputHint})
		}
		if !reflect.DeepEqual(oldV["input_tip"], newV["input_tip"]) {
			actions = append(
				actions,
				producttypes.SetAttributeDefinitionInputTip{
					AttributeName: name, InputTip: attrDef.InputTip})
		}
		if oldV["constraint"] != newV["constraint"] {
			actions = append(
				actions,
				producttypes.ChangeAttributeConstraint{
					AttributeName: name, NewValue: attrDef.AttributeConstraint})
		}

		newFieldType := attrDef.Type
		oldFieldType := oldV["type"].([]interface{})[0].(map[string]interface{})
		oldEnumKeys := make(map[string]interface{})
		newEnumKeys := make(map[string]interface{})

		if enumType, ok := newFieldType.(producttypes.EnumType); ok {
			oldEnumV := oldFieldType["values"].(map[string]interface{})

			for _, enumValue := range enumType.Values {
				newEnumKeys[enumValue.Key] = enumValue
				if _, ok := oldEnumV[enumValue.Key]; !ok {
					// Key does not appear in old enum values, so we'll add it
					actions = append(
						actions,
						producttypes.AddPlainEnumValue{
							AttributeName: name,
							Value:         enumValue,
						})
				}
			}

			// Action: changePlainEnumValueOrder
			// TODO: Change the order of EnumValues: https://docs.commercetools.com/http-api-projects-productTypes.html#change-the-order-of-enumvalues

		} else if enumType, ok := newFieldType.(producttypes.LocalizedEnumType); ok {
			oldEnumV := oldFieldType["localized_value"].([]interface{})

			for _, value := range oldEnumV {
				v := value.(map[string]interface{})
				oldEnumKeys[v["key"].(string)] = v
			}

			for _, enumValue := range enumType.Values {
				newEnumKeys[enumValue.Key] = enumValue
				if _, ok := oldEnumKeys[enumValue.Key]; !ok {
					// Key does not appear in old enum values, so we'll add it
					actions = append(
						actions,
						producttypes.AddLocalizedEnumValue{
							AttributeName: name,
							Value:         enumValue,
						})
				}
			}

			// Action: changeLocalizedEnumValueOrder
			// TODO: Change the order of LocalizedEnumValues: https://docs.commercetools.com/http-api-projects-productTypes.html#change-the-order-of-localizedenumvalues
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
				producttypes.RemoveEnumValues{
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
			producttypes.ChangeAttributeDefinitionsOrder{
				Attributes: newAttrDefinitions,
			})
	}

	log.Printf("[DEBUG] Construction Attribute change actions")

	return actions, nil
}

func resourceProductTypeGetAttributeDefinitions(d *schema.ResourceData) ([]producttypes.AttributeDefinitionDraft, error) {
	input := d.Get("attribute").([]interface{})
	var result []producttypes.AttributeDefinitionDraft

	for _, raw := range input {
		fieldDef, err := resourceProductTypeGetAttributeDefinition(raw.(map[string]interface{}), true)

		if err != nil {
			return nil, err
		}

		result = append(result, fieldDef.(producttypes.AttributeDefinitionDraft))
	}

	return result, nil
}

func resourceProductTypeGetAttributeDefinition(input map[string]interface{}, draft bool) (interface{}, error) {
	attrTypes := input["type"].([]interface{})
	attrType, err := getAttributeType(attrTypes[0])
	if err != nil {
		return nil, err
	}

	label := commercetools.LocalizedString(
		expandStringMap(input["label"].(map[string]interface{})))

	var inputTip commercetools.LocalizedString
	if inputTipRaw, ok := input["input_tip"]; ok {
		inputTip = commercetools.LocalizedString(
			expandStringMap(inputTipRaw.(map[string]interface{})))
	}

	constraint := producttypes.NoneAttributeConstraint
	constraint, ok := constraintMap[input["constraint"].(string)]
	if !ok {
		constraint = producttypes.NoneAttributeConstraint
	}

	if draft {
		return producttypes.AttributeDefinitionDraft{
			Type:                attrType,
			Name:                input["name"].(string),
			Label:               label,
			AttributeConstraint: constraint,
			IsRequired:          input["required"].(bool),
			IsSearchable:        input["searchable"].(bool),
			InputHint:           commercetools.TextInputHint(input["input_hint"].(string)),
			InputTip:            inputTip,
		}, nil
	}
	return producttypes.AttributeDefinition{
		Type:                attrType,
		Name:                input["name"].(string),
		Label:               label,
		AttributeConstraint: constraint,
		IsRequired:          input["required"].(bool),
		IsSearchable:        input["searchable"].(bool),
		InputHint:           commercetools.TextInputHint(input["input_hint"].(string)),
		InputTip:            inputTip,
	}, nil
}

func getAttributeType(input interface{}) (producttypes.AttributeType, error) {
	config := input.(map[string]interface{})
	typeName, ok := config["name"].(string)

	if !ok {
		return nil, fmt.Errorf("No 'name' for type object given")
	}

	switch typeName {
	case "boolean":
		return producttypes.BooleanType{}, nil
	case "text":
		return producttypes.TextType{}, nil
	case "ltext":
		return producttypes.LocalizedTextType{}, nil
	case "enum":
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
		return producttypes.EnumType{Values: values}, nil
	case "lenum":
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
		return producttypes.LocalizedEnumType{Values: values}, nil
	case "number":
		return producttypes.NumberType{}, nil
	case "money":
		return producttypes.MoneyType{}, nil
	case "date":
		return producttypes.DateType{}, nil
	case "time":
		return producttypes.TimeType{}, nil
	case "datetime":
		return producttypes.DateTimeType{}, nil
	case "reference":
		refTypeID, refTypeIDOk := config["reference_type_id"].(string)
		if !refTypeIDOk {
			return nil, fmt.Errorf("No reference_type_id specified for Reference type")
		}
		return producttypes.ReferenceType{
			ReferenceTypeID: refTypeID,
		}, nil
	case "nested":
		typeReferenceID, typeReferenceIDOk := config["reference_type_id"].(string)
		if !typeReferenceIDOk {
			return nil, fmt.Errorf("No type_reference_id specified for Nested type")
		}
		return producttypes.NestedType{
			TypeReference: commercetools.Reference{ID: typeReferenceID, TypeID: "product-type"},
		}, nil
	case "set":
		elementTypes, elementTypesOk := config["element_type"]
		if !elementTypesOk {
			return nil, fmt.Errorf("No element_type specified for Set type")
		}
		elementTypeList := elementTypes.([]interface{})
		if len(elementTypeList) == 0 {
			return nil, fmt.Errorf("No element_type specified for Set type")
		}

		setAttrType, err := getAttributeType(elementTypeList[0])
		if err != nil {
			return nil, err
		}

		return producttypes.SetType{
			ElementType: setAttrType,
		}, nil
	}

	return nil, fmt.Errorf("Unkown AttributeType %s", typeName)
}

func getProductTypeService(m interface{}) *producttypes.Service {
	client := m.(*commercetools.Client)
	svc := producttypes.New(client)
	return svc
}
