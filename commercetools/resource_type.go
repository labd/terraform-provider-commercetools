package commercetools

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v2"
	"github.com/elliotchance/pie/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceTypeResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateTypeStateV0toV1,
				Version: 0,
			},
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
		CustomizeDiff: customdiff.ValidateChange("field", func(ctx context.Context, old, new, meta any) error {
			return resourceTypeValidateField(old.([]any), new.([]any))
		}),
	}
}

func resourceTypeCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))

	resourceTypeIds := []platform.ResourceTypeId{}
	for _, item := range expandStringArray(d.Get("resource_type_ids").([]any)) {
		resourceTypeIds = append(resourceTypeIds, platform.ResourceTypeId(item))

	}
	fields, err := expandTypeFieldDefinition(d)

	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	draft := platform.TypeDraft{
		Key:              d.Get("key").(string),
		Name:             name,
		Description:      &description,
		ResourceTypeIds:  resourceTypeIds,
		FieldDefinitions: fields,
	}

	var ctType *platform.Type
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error

		ctType, err = client.Types().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	d.SetId(ctType.ID)
	d.Set("version", ctType.Version)

	return resourceTypeRead(ctx, d, m)
}

func resourceTypeRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	ctType, err := client.Types().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if utils.IsResourceNotFoundError(err) {
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
		d.Set("resource_type_ids", ctType.ResourceTypeIds)

		if fields, err := flattenTypeFields(ctType); err == nil {
			d.Set("field", fields)
		} else {
			return diag.FromErr(err)
		}
	}
	return nil
}

func resourceTypeUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
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
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.TypeChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.TypeSetDescriptionAction{
				Description: &newDescription})
	}

	if d.HasChange("field") {
		old, new := d.GetChange("field")
		fieldChangeActions, err := resourceTypeFieldChangeActions(old.([]any), new.([]any))
		if err != nil {
			// Workaround invalid state to be written, see
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
			d.Partial(true)
			return diag.FromErr(err)
		}
		input.Actions = append(input.Actions, fieldChangeActions...)
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.Types().WithId(d.Id()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceTypeRead(ctx, d, m)
}
func resourceTypeDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.Types().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	return diag.FromErr(err)
}

func resourceTypeValidateField(old, new []any) error {
	oldLookup := createLookup(old, "name")

	for _, field := range new {
		newF := field.(map[string]any)
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
				"field '%s' type changed from %s to %s."+
					" Changing types is not supported;"+
					" please remove the field first and re-define it later",
				name, oldTypeName, newTypeName)
		}

		if strings.EqualFold(newTypeName, "Set") {
			oldElement := elementFromSlice(oldType, "element_type")
			newElement := elementFromSlice(newType, "element_type")
			oldElementName := oldElement["name"].(string)
			newElementName := newElement["name"].(string)

			if oldElementName != newElementName {
				return fmt.Errorf(
					"field '%s' element type changed from %s to %s."+
						" Changing element types is not supported;"+
						" please remove the field first and re-define it later",
					name, oldElementName, newElementName)
			}
		}

		if oldF["required"] != newF["required"] {
			return fmt.Errorf(
				"error on the '%s' field: "+
					"Updating the 'required' attribute is not supported."+
					"Consider removing the field first and then re-adding it",
				name)
		}
	}
	return nil
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

func valueElement() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"label": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func fieldTypeElement(setsAllowed bool) *schema.Resource {
	result := map[string]*schema.Schema{
		"name": {
			Type: schema.TypeString,
			Description: "Name of the field type. Some types require extra " +
				"fields to be set. Note that changing the type after creating is " +
				"not supported. You need to delete the attribute and re-add it.",
			Required: true,
			ValidateFunc: func(val any, key string) (warns []string, errs []error) {
				v := val.(string)
				validValues := []string{
					"Boolean",
					"Number",
					"String",
					"LocalizedString",
					"Enum",
					"LocalizedEnum",
					"Money",
					"Date",
					"Time",
					"DateTime",
					"Reference",
					"Set",
				}

				if !pie.Contains(validValues, v) {
					errs = append(errs, fmt.Errorf("%s is not a valid type. Valid types are: %s",
						v, strings.Join(pie.SortStableUsing(validValues, func(a, b string) bool {
							return a < b
						}), ", ")))
				}

				if !setsAllowed && v == "Set" {
					errs = append(errs, fmt.Errorf("sets in another Set are not allowed"))
				}
				return
			},
		},
		"value": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Values for the `enum` type.",
			Elem:        valueElement(),
		},
		"localized_value": {
			Type:        schema.TypeList,
			Optional:    true,
			Description: "Localized values for the `lenum` type.",
			Elem:        localizedValueElement(),
		},
		"reference_type_id": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Resource type the Custom Field can reference. Required when type is `Reference`",
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

func expandTypeFieldDefinition(d *schema.ResourceData) ([]platform.FieldDefinition, error) {
	input := d.Get("field").([]any)
	result := make([]platform.FieldDefinition, len(input))

	for i := range input {
		raw := input[i].(map[string]any)
		fieldDef, err := expandTypeFieldDefinitionItem(raw)

		if err != nil {
			return nil, err
		}

		result[i] = *fieldDef
	}

	return result, nil
}

func expandTypeFieldDefinitionItem(input map[string]any) (*platform.FieldDefinition, error) {
	fieldData := elementFromSlice(input, "type")
	if fieldData == nil {
		return nil, fmt.Errorf("missing type")
	}

	fieldType, err := expandTypeFieldType(fieldData)
	if err != nil {
		return nil, err
	}

	label := expandLocalizedString(input["label"])
	inputHint := platform.TypeTextInputHint(input["input_hint"].(string))
	result := &platform.FieldDefinition{
		Type:      fieldType,
		Name:      input["name"].(string),
		Label:     label,
		Required:  input["required"].(bool),
		InputHint: &inputHint,
	}
	return result, nil
}

func expandTypeFieldType(input any) (platform.FieldType, error) {
	config := input.(map[string]any)
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
		valuesInput, valuesOk := config["value"]
		if !valuesOk {
			return nil, fmt.Errorf("no value elements specified for Enum type")
		}
		var values []platform.CustomFieldEnumValue
		for _, value := range valuesInput.([]any) {
			v := value.(map[string]any)
			values = append(values, platform.CustomFieldEnumValue{
				Key:   v["key"].(string),
				Label: v["label"].(string),
			})
		}
		return platform.CustomFieldEnumType{Values: values}, nil
	case "LocalizedEnum":
		valuesInput, valuesOk := config["localized_value"]
		if !valuesOk {
			return nil, fmt.Errorf("no localized_value elements specified for LocalizedEnum type")
		}
		var values []platform.CustomFieldLocalizedEnumValue
		for _, value := range valuesInput.([]any) {
			v := value.(map[string]any)
			labels := expandLocalizedString(v["label"])
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
		if ref, ok := config["reference_type_id"].(string); ok {
			result := platform.CustomFieldReferenceType{
				ReferenceTypeId: platform.CustomFieldReferenceValue(ref),
			}
			return result, nil
		}
		return nil, fmt.Errorf("no reference_type_id specified for Reference type")
	case "Set":
		data := elementFromSlice(config, "element_type")
		if data == nil {
			return nil, fmt.Errorf("no element_type specified for Set type")
		}

		setFieldType, err := expandTypeFieldType(data)
		if err != nil {
			return nil, err
		}
		result := platform.CustomFieldSetType{
			ElementType: setFieldType,
		}
		return result, nil
	}

	return nil, fmt.Errorf("unknown FieldType %s", typeName)
}

func flattenTypeFields(t *platform.Type) ([]map[string]any, error) {
	fields := make([]map[string]any, len(t.FieldDefinitions))
	for i, fieldDef := range t.FieldDefinitions {
		fieldType, err := flattenTypeFieldType(fieldDef.Type, true)
		if err != nil {
			return nil, err
		}
		fields[i] = map[string]any{
			"type":       fieldType,
			"name":       fieldDef.Name,
			"label":      fieldDef.Label,
			"required":   fieldDef.Required,
			"input_hint": fieldDef.InputHint,
		}
	}
	return fields, nil
}

func flattenTypeFieldType(fieldType platform.FieldType, setsAllowed bool) ([]any, error) {
	typeData := make(map[string]any)

	switch val := fieldType.(type) {

	case platform.CustomFieldBooleanType:
		typeData["name"] = "Boolean"

	case platform.CustomFieldStringType:
		typeData["name"] = "String"

	case platform.CustomFieldLocalizedStringType:
		typeData["name"] = "LocalizedString"

	case platform.CustomFieldEnumType:
		typeData["name"] = "Enum"
		typeData["value"] = flattenTypePlainEnum(val.Values)
	case platform.CustomFieldLocalizedEnumType:
		typeData["name"] = "LocalizedEnum"
		typeData["localized_value"] = flattenTypeLocalizedEnum(val.Values)

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
			elemType, err := flattenTypeFieldType(val.ElementType, false)
			if err != nil {
				return nil, err
			}
			typeData["element_type"] = elemType
		}

	default:
		return nil, fmt.Errorf("unknown resource Type %T: %#v", fieldType, fieldType)
	}

	return []any{typeData}, nil
}

func flattenTypeLocalizedEnum(values []platform.CustomFieldLocalizedEnumValue) []any {
	enumValues := make([]any, len(values))
	for i := range values {
		enumValues[i] = map[string]any{
			"key":   values[i].Key,
			"label": values[i].Label,
		}
	}
	return enumValues
}

func mapFieldDefinition(values []any) (*orderedmap.OrderedMap[string, platform.FieldDefinition], error) {
	fields := orderedmap.NewOrderedMap[string, platform.FieldDefinition]()
	for i := range values {
		raw := values[i].(map[string]any)
		field, err := expandTypeFieldDefinitionItem(raw)
		if err != nil {
			return nil, err
		}

		fields.Set(field.Name, *field)
	}
	return fields, nil
}

// Generate a list of actions needed for updating the fields value in
// commercetools so that it matches the terraform file
func resourceTypeFieldChangeActions(oldValues []any, newValues []any) ([]platform.TypeUpdateAction, error) {
	oldFields, err := mapFieldDefinition(oldValues)
	if err != nil {
		return nil, err
	}

	newFields, err := mapFieldDefinition(newValues)
	if err != nil {
		return nil, err
	}

	// Create a copy of the field order for commercetools. When we
	// delete fields commercetools already re-orders the fields and we need
	// to not send a reorder command when the order already matches
	fieldOrder := []string{}
	fieldOrder = append(fieldOrder, oldFields.Keys()...)

	actions := []platform.TypeUpdateAction{}

	// Check if we have fields which are removed and generate the corresponding
	// remove field actions
	for _, name := range oldFields.Keys() {
		if _, ok := newFields.Get(name); !ok {
			actions = append(actions, platform.TypeRemoveFieldDefinitionAction{FieldName: name})
			fieldOrder = removeValueFromSlice(fieldOrder, name)
		}
	}

	for _, name := range newFields.Keys() {
		newField, _ := newFields.Get(name)
		oldField, isExisting := oldFields.Get(name)

		// A new field is added. Create the update action skip the rest of the
		// loop since there cannot be any change if the field didn't exist yet.
		if !isExisting {
			actions = append(
				actions,
				platform.TypeAddFieldDefinitionAction{
					FieldDefinition: newField,
				})
			fieldOrder = append(fieldOrder, newField.Name)
			continue
		}

		// This should not be able to happen due to checks earlier
		if reflect.TypeOf(oldField.Type) != reflect.TypeOf(newField.Type) {
			return nil, fmt.Errorf("changing field types is not supported in commercetools")
		}

		// Check if we need to update the field label
		if !reflect.DeepEqual(oldField.Label, newField.Label) {
			actions = append(
				actions,
				platform.TypeChangeLabelAction{
					FieldName: name,
					Label:     newField.Label,
				})
		}

		// Update the input hint if this is changed
		if !reflect.DeepEqual(oldField.InputHint, newField.InputHint) {
			actions = append(
				actions,
				platform.TypeChangeInputHintAction{
					FieldName: name,
					InputHint: *newField.InputHint,
				})
		}

		// Specific updates for EnumType, LocalizedEnumType and a Set of these
		switch t := newField.Type.(type) {

		case platform.CustomFieldLocalizedEnumType:
			ot := oldField.Type.(platform.CustomFieldLocalizedEnumType)
			subActions, err := updateCustomFieldLocalizedEnumType(name, ot, t)
			if err != nil {
				return nil, err
			}
			actions = append(actions, subActions...)

		case platform.CustomFieldEnumType:
			ot := oldField.Type.(platform.CustomFieldEnumType)
			subActions, err := updateCustomFieldEnumType(name, ot, t)
			if err != nil {
				return nil, err
			}
			actions = append(actions, subActions...)

		case platform.CustomFieldSetType:
			ot := oldField.Type.(platform.CustomFieldSetType)

			// This should not be able to happen due to checks earlier
			if reflect.TypeOf(ot.ElementType) != reflect.TypeOf(t.ElementType) {
				return nil, fmt.Errorf("changing field types is not supported in commercetools")
			}

			switch st := t.ElementType.(type) {

			case platform.CustomFieldEnumType:
				ost := ot.ElementType.(platform.CustomFieldEnumType)
				subActions, err := updateCustomFieldEnumType(name, ost, st)
				if err != nil {
					return nil, err
				}
				actions = append(actions, subActions...)

			case platform.CustomFieldLocalizedEnumType:
				ost := ot.ElementType.(platform.CustomFieldLocalizedEnumType)
				subActions, err := updateCustomFieldLocalizedEnumType(name, ost, st)
				if err != nil {
					return nil, err
				}
				actions = append(actions, subActions...)
			}
		}

	}

	if !reflect.DeepEqual(fieldOrder, newFields.Keys()) {
		actions = append(
			actions,
			platform.TypeChangeFieldDefinitionOrderAction{
				FieldNames: newFields.Keys(),
			})
	}

	return actions, nil
}

func updateCustomFieldEnumType(fieldName string, old, new platform.CustomFieldEnumType) ([]platform.TypeUpdateAction, error) {
	oldValues := orderedmap.NewOrderedMap[string, platform.CustomFieldEnumValue]()
	for i := range old.Values {
		oldValues.Set(old.Values[i].Key, old.Values[i])
	}

	newValues := orderedmap.NewOrderedMap[string, platform.CustomFieldEnumValue]()
	for i := range new.Values {
		newValues.Set(new.Values[i].Key, new.Values[i])
	}

	valueOrder := []string{}
	valueOrder = append(valueOrder, oldValues.Keys()...)

	actions := []platform.TypeUpdateAction{}
	for _, key := range newValues.Keys() {
		newValue, _ := newValues.Get(key)

		// Check if this is a new value
		if _, ok := oldValues.Get(key); !ok {
			actions = append(
				actions,
				platform.TypeAddEnumValueAction{
					FieldName: fieldName,
					Value:     newValue,
				})
			valueOrder = append(valueOrder, newValue.Key)
			continue
		}

		oldValue, _ := oldValues.Get(key)

		// Check if the label is changed and create an update action
		if !reflect.DeepEqual(oldValue.Label, newValue.Label) {
			actions = append(
				actions,
				platform.TypeChangeEnumValueLabelAction{
					FieldName: fieldName,
					Value:     newValue,
				})
		}
	}

	// Check if the order is changed. We compare this against valueOrder to take
	// into account new fields added to the end by commercetools
	if !reflect.DeepEqual(valueOrder, newValues.Keys()) {
		actions = append(
			actions,
			platform.TypeChangeEnumValueOrderAction{
				FieldName: fieldName,
				Keys:      newValues.Keys(),
			})

	}

	return actions, nil
}

func updateCustomFieldLocalizedEnumType(fieldName string, old, new platform.CustomFieldLocalizedEnumType) ([]platform.TypeUpdateAction, error) {
	oldValues := orderedmap.NewOrderedMap[string, platform.CustomFieldLocalizedEnumValue]()
	for i := range old.Values {
		oldValues.Set(old.Values[i].Key, old.Values[i])
	}

	newValues := orderedmap.NewOrderedMap[string, platform.CustomFieldLocalizedEnumValue]()
	for i := range new.Values {
		newValues.Set(new.Values[i].Key, new.Values[i])
	}

	valueOrder := []string{}
	valueOrder = append(valueOrder, oldValues.Keys()...)

	actions := []platform.TypeUpdateAction{}
	for _, key := range newValues.Keys() {
		newValue, _ := newValues.Get(key)

		// Check if this is a new value
		if _, ok := oldValues.Get(key); !ok {
			actions = append(
				actions,
				platform.TypeAddLocalizedEnumValueAction{
					FieldName: fieldName,
					Value:     newValue,
				})
			valueOrder = append(valueOrder, newValue.Key)
			continue
		}

		oldValue, _ := oldValues.Get(key)

		// Check if the label is changed and create an update action
		if !reflect.DeepEqual(oldValue.Label, newValue.Label) {
			actions = append(
				actions,
				platform.TypeChangeLocalizedEnumValueLabelAction{
					FieldName: fieldName,
					Value:     newValue,
				})
		}
	}

	// Check if the order is changed. We compare this against valueOrder to take
	// into account new fields added to the end by commercetools
	if !reflect.DeepEqual(valueOrder, newValues.Keys()) {
		actions = append(
			actions,
			platform.TypeChangeLocalizedEnumValueOrderAction{
				FieldName: fieldName,
				Keys:      newValues.Keys(),
			})

	}

	return actions, nil
}

func flattenTypePlainEnum(values []platform.CustomFieldEnumValue) []any {
	enumValues := make([]any, len(values))
	for i, value := range values {
		enumValues[i] = map[string]any{
			"key":   value.Key,
			"label": value.Label,
		}
	}
	return enumValues
}
