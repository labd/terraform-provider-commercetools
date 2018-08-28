package commercetools

import (
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
	"github.com/labd/commercetools-go-sdk/cterrors"
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
						},
					},
				},
			},
			"version": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
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
			Elem: &schema.Resource{
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
			},
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
			if reqerr, ok := err.(cterrors.RequestError); ok {
				log.Printf("[DEBUG] Received RequestError %s", reqerr)
				if reqerr.StatusCode() == 400 {
					return resource.NonRetryableError(reqerr)
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
		if reqerr, ok := err.(cterrors.RequestError); ok {
			log.Printf("[DEBUG] Received RequestError %s", reqerr)
			if reqerr.StatusCode() == 404 {
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

		// TODO: Implement Read method

		d.Set("version", ctType.Version)
	}
	return nil
}

func resourceTypeUpdate(d *schema.ResourceData, m interface{}) error {
	svc := getTypeService(m)

	input := &types.UpdateInput{
		ID:      d.Id(),
		Version: d.Get("version").(int),
		Actions: commercetools.UpdateActions{},
	}

	// TODO: Implement UpdateActions

	_, err := svc.Update(input)
	if err != nil {
		return err
	}

	return resourceTypeRead(d, m)
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
		i := raw.(map[string]interface{})
		fieldTypes, ok := i["type"].(*schema.Set)

		if !ok {
			return nil, fmt.Errorf("No type defined for field definition")
		}
		if fieldTypes.Len() > 1 {
			return nil, fmt.Errorf("More then 1 type definition detected. Please remove the redundant ones")
		}
		fieldType, err := getFieldType(fieldTypes.List()[0].(map[string]interface{}))
		if err != nil {
			return nil, err
		}

		label := commercetools.LocalizedString(
			expandStringMap(i["label"].(map[string]interface{})))

		result = append(result, types.FieldDefinition{
			Type:      fieldType,
			Name:      i["name"].(string),
			Label:     label,
			Required:  i["required"].(bool),
			InputHint: commercetools.TextInputHint(i["input_hint"].(string)),
		})
	}

	return result, nil
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
