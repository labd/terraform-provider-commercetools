package commercetools

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/elliotchance/orderedmap/v2"
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
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceProductTypeResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateProductTypeStateV0toV1,
				Version: 0,
			},
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
			Type: schema.TypeString,
			Description: "Name of the field type. Some types require extra " +
				"fields to be set. Note that changing the type after creating is " +
				"not supported. You need to delete the attribute and re-add it",
			Required: true,
			ValidateFunc: func(val any, key string) (warns []string, errs []error) {
				v := val.(string)
				if !setsAllowed && v == "set" {
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
			Description: "Resource type the Custom Field can reference. Required when type is `reference`",
		},
		"type_reference": {
			Type:        schema.TypeString,
			Description: "Reference to another product type. Required when type is `nested`.",
			Optional:    true,
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
	attrs := make([]map[string]any, len(t.Attributes))
	for i, attrDef := range t.Attributes {
		attrType, err := flattenProductTypeAttributeType(attrDef.Type, true)
		if err != nil {
			return nil, err
		}
		attrs[i] = map[string]any{
			"type":       attrType,
			"name":       attrDef.Name,
			"label":      attrDef.Label,
			"required":   attrDef.IsRequired,
			"input_hint": attrDef.InputHint,
			"constraint": attrDef.AttributeConstraint,
			"searchable": attrDef.IsSearchable,
		}
		if attrDef.InputTip != nil {
			attrs[i]["input_tip"] = *attrDef.InputTip
		}
	}
	return attrs, nil
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
		typeData["name"] = "enum"
		typeData["value"] = flattenProductTypePlainEnum(f.Values)
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
		attrChangeActions, err := resourceProductTypeAttributeChangeActions(
			old.([]any), new.([]any))
		if err != nil {
			return diag.FromErr(err)
		}
		input.Actions = append(input.Actions, attrChangeActions...)
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
	oldAttrs, err := mapAttributeDefinition(oldValues)
	if err != nil {
		return nil, err
	}

	newAttrs, err := mapAttributeDefinition(newValues)
	if err != nil {
		return nil, err
	}

	// Create a copy of the attribute order for commercetools. When we
	// delete attributes commercetools already re-orders the attributes and we need
	// to not send a reorder command when the order already matches
	attrOrder := []string{}
	attrOrder = append(attrOrder, oldAttrs.Keys()...)

	actions := []platform.ProductTypeUpdateAction{}

	// Check if we have attributes which are removed and generate the corresponding
	// remove attribute actions
	for _, name := range oldAttrs.Keys() {
		if _, ok := newAttrs.Get(name); !ok {
			actions = append(actions, platform.ProductTypeRemoveAttributeDefinitionAction{Name: name})
			attrOrder = removeValueFromSlice(attrOrder, name)
		}
	}

	for _, name := range newAttrs.Keys() {
		newAttr, _ := newAttrs.Get(name)
		oldAttr, isExisting := oldAttrs.Get(name)

		// A new attribute is added. Create the update action skip the rest of the
		// loop since there cannot be any change if the attribute didn't exist yet.
		if !isExisting {
			actions = append(
				actions,
				platform.ProductTypeAddAttributeDefinitionAction{
					Attribute: platform.AttributeDefinitionDraft{
						Type:                newAttr.Type,
						Name:                newAttr.Name,
						Label:               newAttr.Label,
						IsRequired:          newAttr.IsRequired,
						AttributeConstraint: &newAttr.AttributeConstraint,
						InputTip:            newAttr.InputTip,
						InputHint:           &newAttr.InputHint,
						IsSearchable:        &newAttr.IsSearchable,
					},
				})
			attrOrder = append(attrOrder, newAttr.Name)
			continue
		}

		// This should not be able to happen due to checks earlier
		if reflect.TypeOf(oldAttr.Type) != reflect.TypeOf(newAttr.Type) {
			return nil, fmt.Errorf("changing attribute types is not supported in commercetools")
		}

		// Check if we need to update the attribute label
		if !reflect.DeepEqual(oldAttr.Label, newAttr.Label) {
			actions = append(
				actions,
				platform.ProductTypeChangeLabelAction{
					AttributeName: name,
					Label:         newAttr.Label,
				})
		}

		if !reflect.DeepEqual(oldAttr.Name, newAttr.Name) {
			actions = append(
				actions,
				platform.ProductTypeChangeNameAction{
					Name: newAttr.Name,
				})
		}

		if !reflect.DeepEqual(oldAttr.IsSearchable, newAttr.IsSearchable) {
			actions = append(
				actions,
				platform.ProductTypeChangeIsSearchableAction{
					AttributeName: name,
					IsSearchable:  newAttr.IsSearchable,
				})
		}

		// Update the input hint if this is changed
		if !reflect.DeepEqual(oldAttr.InputHint, newAttr.InputHint) {
			actions = append(
				actions,
				platform.ProductTypeChangeInputHintAction{
					AttributeName: name,
					NewValue:      newAttr.InputHint,
				})
		}

		if !reflect.DeepEqual(oldAttr.InputTip, newAttr.InputTip) {
			actions = append(
				actions,
				platform.ProductTypeSetInputTipAction{
					AttributeName: name,
					InputTip:      newAttr.InputTip,
				})
		}

		if !reflect.DeepEqual(oldAttr.AttributeConstraint, newAttr.AttributeConstraint) {
			actions = append(
				actions,
				platform.ProductTypeChangeAttributeConstraintAction{
					AttributeName: name,
					NewValue:      platform.AttributeConstraintEnumDraft(newAttr.AttributeConstraint),
				})
		}

		// Specific updates for EnumType, LocalizedEnumType and a Set of these
		switch t := newAttr.Type.(type) {

		case platform.AttributeLocalizedEnumType:
			ot := oldAttr.Type.(platform.AttributeLocalizedEnumType)
			subActions, err := updateAttributeLocalizedEnumType(name, ot, t)
			if err != nil {
				return nil, err
			}
			actions = append(actions, subActions...)

		case platform.AttributeEnumType:
			ot := oldAttr.Type.(platform.AttributeEnumType)
			subActions, err := updateAttributeEnumType(name, ot, t)
			if err != nil {
				return nil, err
			}
			actions = append(actions, subActions...)

		case platform.AttributeSetType:
			ot := oldAttr.Type.(platform.AttributeSetType)

			// This should not be able to happen due to checks earlier
			if reflect.TypeOf(ot.ElementType) != reflect.TypeOf(t.ElementType) {
				return nil, fmt.Errorf("changing attribute types is not supported in commercetools")
			}

			switch st := t.ElementType.(type) {

			case platform.AttributeEnumType:
				ost := ot.ElementType.(platform.AttributeEnumType)
				subActions, err := updateAttributeEnumType(name, ost, st)
				if err != nil {
					return nil, err
				}
				actions = append(actions, subActions...)

			case platform.AttributeLocalizedEnumType:
				ost := ot.ElementType.(platform.AttributeLocalizedEnumType)
				subActions, err := updateAttributeLocalizedEnumType(name, ost, st)
				if err != nil {
					return nil, err
				}
				actions = append(actions, subActions...)
			}
		}

	}

	if !reflect.DeepEqual(attrOrder, newAttrs.Keys()) {
		attrs := make([]platform.AttributeDefinition, newAttrs.Len())
		for i, key := range newAttrs.Keys() {
			if el, ok := newAttrs.Get(key); ok {
				attrs[i] = el
			}
		}
		actions = append(
			actions,
			platform.ProductTypeChangeAttributeOrderAction{
				Attributes: attrs,
			})
	}

	return actions, nil
}

func updateAttributeEnumType(attrName string, old, new platform.AttributeEnumType) ([]platform.ProductTypeUpdateAction, error) {
	oldValues := orderedmap.NewOrderedMap[string, platform.AttributePlainEnumValue]()
	for i := range old.Values {
		oldValues.Set(old.Values[i].Key, old.Values[i])
	}

	newValues := orderedmap.NewOrderedMap[string, platform.AttributePlainEnumValue]()
	for i := range new.Values {
		newValues.Set(new.Values[i].Key, new.Values[i])
	}

	valueOrder := []string{}
	valueOrder = append(valueOrder, oldValues.Keys()...)

	actions := []platform.ProductTypeUpdateAction{}

	// Delete enum values
	removeKeys := []string{}
	for _, key := range oldValues.Keys() {
		if _, ok := newValues.Get(key); !ok {
			removeKeys = append(removeKeys, key)
			valueOrder = removeValueFromSlice(valueOrder, key)
		}
	}
	if len(removeKeys) > 0 {
		actions = append(
			actions,
			platform.ProductTypeRemoveEnumValuesAction{
				AttributeName: attrName,
				Keys:          removeKeys,
			})
	}

	for _, key := range newValues.Keys() {
		newValue, _ := newValues.Get(key)

		// Check if this is a new value
		if _, ok := oldValues.Get(key); !ok {
			actions = append(
				actions,
				platform.ProductTypeAddPlainEnumValueAction{
					AttributeName: attrName,
					Value:         newValue,
				})
			valueOrder = append(valueOrder, newValue.Key)
			continue
		}

		oldValue, _ := oldValues.Get(key)

		// Check if the label is changed and create an update action
		if !reflect.DeepEqual(oldValue.Label, newValue.Label) {
			actions = append(
				actions,
				platform.ProductTypeChangePlainEnumValueLabelAction{
					AttributeName: attrName,
					NewValue:      newValue,
				})
		}
	}

	// Check if the order is changed. We compare this against valueOrder to take
	// into account new attributes added to the end by commercetools
	if !reflect.DeepEqual(valueOrder, newValues.Keys()) {
		values := make([]platform.AttributePlainEnumValue, newValues.Len())
		for i, key := range newValues.Keys() {
			if el, ok := newValues.Get(key); ok {
				values[i] = el
			}
		}
		actions = append(
			actions,
			platform.ProductTypeChangePlainEnumValueOrderAction{
				AttributeName: attrName,
				Values:        values,
			})
	}

	return actions, nil
}

func updateAttributeLocalizedEnumType(attrName string, old, new platform.AttributeLocalizedEnumType) ([]platform.ProductTypeUpdateAction, error) {
	oldValues := orderedmap.NewOrderedMap[string, platform.AttributeLocalizedEnumValue]()
	for i := range old.Values {
		oldValues.Set(old.Values[i].Key, old.Values[i])
	}

	newValues := orderedmap.NewOrderedMap[string, platform.AttributeLocalizedEnumValue]()
	for i := range new.Values {
		newValues.Set(new.Values[i].Key, new.Values[i])
	}

	valueOrder := []string{}
	valueOrder = append(valueOrder, oldValues.Keys()...)

	actions := []platform.ProductTypeUpdateAction{}

	// Delete enum values
	removeKeys := []string{}
	for _, key := range oldValues.Keys() {
		if _, ok := newValues.Get(key); !ok {
			removeKeys = append(removeKeys, key)
			valueOrder = removeValueFromSlice(valueOrder, key)
		}
	}
	if len(removeKeys) > 0 {
		actions = append(
			actions,
			platform.ProductTypeRemoveEnumValuesAction{
				AttributeName: attrName,
				Keys:          removeKeys,
			})
	}

	for _, key := range newValues.Keys() {
		newValue, _ := newValues.Get(key)

		// Check if this is a new value
		if _, ok := oldValues.Get(key); !ok {
			actions = append(
				actions,
				platform.ProductTypeAddLocalizedEnumValueAction{
					AttributeName: attrName,
					Value:         newValue,
				})
			valueOrder = append(valueOrder, newValue.Key)
			continue
		}

		oldValue, _ := oldValues.Get(key)

		// Check if the label is changed and create an update action
		if !reflect.DeepEqual(oldValue.Label, newValue.Label) {
			actions = append(
				actions,
				platform.ProductTypeChangeLocalizedEnumValueLabelAction{
					AttributeName: attrName,
					NewValue:      newValue,
				})
		}
	}

	// Check if the order is changed. We compare this against valueOrder to take
	// into account new attributes added to the end by commercetools
	if !reflect.DeepEqual(valueOrder, newValues.Keys()) {
		values := make([]platform.AttributeLocalizedEnumValue, newValues.Len())
		for i, key := range newValues.Keys() {
			if el, ok := newValues.Get(key); ok {
				values[i] = el
			}
		}
		actions = append(
			actions,
			platform.ProductTypeChangeLocalizedEnumValueOrderAction{
				AttributeName: attrName,
				Values:        values,
			})
	}

	return actions, nil
}

func mapAttributeDefinition(values []any) (*orderedmap.OrderedMap[string, platform.AttributeDefinition], error) {
	attrs := orderedmap.NewOrderedMap[string, platform.AttributeDefinition]()
	for i := range values {
		raw := values[i].(map[string]any)
		attr, err := expandProductTypeAttributeDefinitionItem(raw, false)
		if err != nil {
			return nil, err
		}

		attrs.Set(attr.(platform.AttributeDefinition).Name, attr.(platform.AttributeDefinition))
	}
	return attrs, nil
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
		valuesInput, valuesOk := config["value"]
		if !valuesOk {
			return nil, fmt.Errorf("no value elements specified for Enum type")
		}
		var values []platform.AttributePlainEnumValue
		for _, value := range valuesInput.([]any) {
			v := value.(map[string]any)
			values = append(values, platform.AttributePlainEnumValue{
				Key:   v["key"].(string),
				Label: v["label"].(string),
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

func flattenProductTypePlainEnum(values []platform.AttributePlainEnumValue) []any {
	enumValues := make([]any, len(values))
	for i, value := range values {
		enumValues[i] = map[string]any{
			"key":   value.Key,
			"label": value.Label,
		}
	}
	return enumValues
}

func attributeTypeElementV0(setsAllowed bool) *schema.Resource {
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
		"value": {
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

func resourceProductTypeResourceV0() *schema.Resource {
	return &schema.Resource{
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
				Description: "[Product attribute definition](https://docs.commercetools.com/api/projects/productTypes#attributedefinition)",
				Type:        schema.TypeList,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Description: "[AttributeType](https://docs.commercetools.com/api/projects/productTypes#attributetype)",
							Type:        schema.TypeList,
							MaxItems:    1,
							Required:    true,
							Elem:        attributeTypeElementV0(true),
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
	}
}

func migrateProductTypeStateV0toV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	if attr, ok := rawState["attribute"].([]any); ok {
		// iterate over all attributes
		for _, item := range attr {
			if m, ok := item.(map[string]interface{}); ok {
				// check attribute.type
				if itemTypes, ok := m["type"].([]any); ok {
					// it should only contain 1 element, which is an array
					if len(itemTypes) == 1 {
						if itemType, ok := itemTypes[0].(map[string]any); ok {
							if itemTypeName, ok := itemType["name"].(string); ok {
								if itemTypeName == "set" {
									if itemTypeElementType, ok := itemType["element_type"].([]any); ok {
										if len(itemTypeElementType) == 1 {
											if itemTypeElementTypeValues, ok := itemTypeElementType[0].(map[string]any)["values"]; ok {
												if itemTypeElementTypeValues, ok := itemTypeElementTypeValues.(map[string]any); ok {
													// "values" and "value" cannot co exist, so this needs an upgrade
													value := make([]map[string]string, len(itemTypeElementTypeValues))
													i := 0
													for _, itemTypeElementTypeValue := range itemTypeElementTypeValues {
														value[i] = map[string]string{
															"key":   itemTypeElementTypeValue.(string),
															"label": itemTypeElementTypeValue.(string),
														}
														i++
													}
													// add "value"
													itemTypeElementType[0].(map[string]any)["value"] = value
													// remove "values"
													delete(itemTypeElementType[0].(map[string]any), "values")
												}
											}
										}
									}
								} else if itemTypeName == "enum" {
									if itemTypeValues, ok := itemType["values"].(map[string]any); ok {
										// "values" and "value" cannot co exist, so this needs an upgrade
										value := make([]map[string]string, len(itemTypeValues))
										i := 0
										for _, itemTypeValue := range itemTypeValues {
											value[i] = map[string]string{
												"key":   itemTypeValue.(string),
												"label": itemTypeValue.(string),
											}
											i++
										}
										// add "value"
										itemType["value"] = value
										// remove "values"
										delete(itemType, "values")
									}
								}
							}
						}
					}
				}
			}
		}
	}
	return rawState, nil
}
