package commercetools

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/labd/terraform-provider-commercetools/internal/utils"
)

func resourceAttributeGroups() *schema.Resource {
	return &schema.Resource{
		Description: "Attribute Groups allow to cluster subsets of Attributes of a Product " +
			"See also the [Attribute Groups Documentation](https://docs.commercetools.com/api/projects/attribute-groups)",
		CreateContext: resourceAttributeGroupsCreate,
		ReadContext:   resourceAttributeGroupsRead,
		UpdateContext: resourceAttributeGroupsUpdate,
		DeleteContext: resourceAttributeGroupsDelete,
		Schema: map[string]*schema.Schema{
			"key": {
				Description: "User-defined unique identifier of the AttributeGroup",
				Type:        schema.TypeString,
				Required:    true,
			},
			"attributes": {
				Description: "Attributes with unique values.",
				Type:        schema.TypeList,
				Required:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"name": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"description": {
				Description:      "[LocalizedString](https://docs.commercetools.com/api/types#localizedstring)",
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceAttributeGroupsCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	name := expandLocalizedString(d.Get("name"))
	description := expandLocalizedString(d.Get("description"))

	attributes := expandStringArray(d.Get("attributes").([]any))

	identifiers := make([]platform.AttributeReference, 0)
	for i := 0; i < len(attributes); i++ {
		channelIdentifier := platform.AttributeReference{
			Key: attributes[i],
		}
		identifiers = append(identifiers, channelIdentifier)
	}

	client := getClient(m)

	draft := platform.AttributeGroupDraft{
		Key:         stringRef(d.Get("key")),
		Attributes:  identifiers,
		Name:        name,
		Description: &description,
	}

	var attributeGroups *platform.AttributeGroup
	var err = resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		var err error

		attributeGroups, err = client.AttributeGroups().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(attributeGroups.ID)
	d.Set("version", attributeGroups.Version)
	return resourceAttributeGroupsRead(ctx, d, m)
}

func resourceAttributeGroupsRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	attributeGroups, err := client.AttributeGroups().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.SetId(attributeGroups.ID)
	d.Set("version", attributeGroups.Version)
	if attributeGroups.Name != nil {
		d.Set("name", attributeGroups.Name)
	}
	if attributeGroups.Description != nil {
		d.Set("description", *attributeGroups.Description)
	} else {
		d.Set("description", nil)
	}
	d.Set("key", attributeGroups.Key)

	if attributeGroups.Attributes != nil {
		attributesKeys, err := flattenAttributes(attributeGroups.Attributes)
		if err != nil {
			return diag.FromErr(err)
		}
		d.Set("attributes", attributesKeys)
	}
	return nil
}

func resourceAttributeGroupsUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.AttributeGroupUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.AttributeGroupUpdateAction{},
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.AttributeGroupSetKeyAction{Key: &newKey})
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.AttributeGroupChangeNameAction{Name: newName})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.AttributeGroupSetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("attributes") {
		attributes := expandStringArray(d.Get("attributes").([]any))

		identifiers := make([]platform.AttributeReference, 0)
		for i := 0; i < len(attributes); i++ {
			channelIdentifier := platform.AttributeReference{
				Key: attributes[i],
			}
			identifiers = append(identifiers, channelIdentifier)
		}
		input.Actions = append(
			input.Actions,
			&platform.AttributeGroupSetAttributesAction{Attributes: identifiers})
	}

	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.AttributeGroups().WithId(d.Id()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceAttributeGroupsRead(ctx, d, m)
}

func resourceAttributeGroupsDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 20*time.Second, func() *resource.RetryError {
		_, err := client.AttributeGroups().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenAttributes(attributes []platform.AttributeReference) ([]string, error) {
	attributeKeys := make([]string, 0)
	for i := 0; i < len(attributes); i++ {
		attributeKeys = append(attributeKeys, attributes[i].Key)
	}
	return attributeKeys, nil
}
