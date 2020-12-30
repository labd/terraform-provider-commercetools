package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/labd/commercetools-go-sdk/commercetools"
)

func resourceCategory() *schema.Resource {
	return &schema.Resource{
		Create: resourceCategoryCreate,
		Read:   resourceCategoryRead,
		Update: resourceCategoryUpdate,
		Delete: resourceCategoryDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeMap,
				Required: true,
			},
			"description": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"slug": {
				Type:     schema.TypeMap,
				Required: true,
			},
			//parent
			"orderHint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"externalId": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"metaTitle": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"metaDescription": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"metaKeywords": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			//custom
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceCategoryCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var category *commercetools.Category

	name := commercetools.LocalizedString(expandStringMap(d.Get("name").(map[string]interface{})))
	desc := commercetools.LocalizedString(expandStringMap(d.Get("description").(map[string]interface{})))
	slug := commercetools.LocalizedString(expandStringMap(d.Get("slug").(map[string]interface{})))
	metaTitle := commercetools.LocalizedString(expandStringMap(d.Get("metaTitle").(map[string]interface{})))
	metaDescription := commercetools.LocalizedString(expandStringMap(d.Get("metaDescription").(map[string]interface{})))
	metaKeywords := commercetools.LocalizedString(expandStringMap(d.Get("metaKeywords").(map[string]interface{})))

	draft := &commercetools.CategoryDraft{
		Key: d.Get("key").(string),
		Name:        &name,
		Description: &desc,
		Slug:        &slug,
		OrderHint:  d.Get("orderHint").(string),
		ExternalID:  d.Get("externalId").(string),
		MetaTitle: &metaTitle,
		MetaDescription: &metaDescription,
		MetaKeywords: &metaKeywords,
	}

	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		var err error

		category, err = client.CategoryCreate(context.Background(), draft)
		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	if category == nil {
		log.Fatal("No  category created?")
	}

	d.SetId(category.ID)
	d.Set("version", category.Version)

	return resourceCategoryRead(d, m)
}

func resourceCategoryRead(d *schema.ResourceData, m interface{}) error {
	log.Printf("[DEBUG] Reading category from commercetools, with category id: %s", d.Id())
	client := getClient(m)

	category, err := client.CategoryGetWithID(context.Background(), d.Id())

	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}

	if category == nil {
		log.Print("[DEBUG] No category found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following category:")
		log.Print(stringFormatObject(category))

		d.Set("version", category.Version)
		d.Set("key", category.Key)
		d.Set("name", *category.Name)
		d.Set("description", *category.Description)
		d.Set("orderHint", category.OrderHint)
		d.Set("externalId", category.ExternalID)
		d.Set("metaTitle", *category.MetaTitle)
		d.Set("metaDescription", *category.MetaDescription)
		d.Set("metaKeywords", *category.MetaKeywords)
	}
	return nil
}

func resourceCategoryUpdate(d *schema.ResourceData, m interface{}) error {
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	client := getClient(m)
	category, err := client.CategoryGetWithID(context.Background(), d.Id())
	if err != nil {
		return err
	}

	input := &commercetools.CategoryUpdateWithIDInput{
		ID:      d.Id(),
		Version: category.Version,
		Actions: []commercetools.CategoryUpdateAction{},
	}

	if d.HasChange("name") {
		newName := commercetools.LocalizedString(expandStringMap(d.Get("name").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategoryChangeNameAction{Name: &newName})
	}

	if d.HasChange("slug") {
		newSlug := commercetools.LocalizedString(expandStringMap(d.Get("slug").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategoryChangeSlugAction{Slug: &newSlug})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetKeyAction{Key: newKey})
	}

	if d.HasChange("orderHint") {
		newVal := d.Get("orderHint").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetKeyAction{Key: newVal})
	}

	if d.HasChange("externalId") {
		newVal := d.Get("externalId").(string)
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetKeyAction{Key: newVal})
	}

	if d.HasChange("description") {
		newDescription := commercetools.LocalizedString(expandStringMap(d.Get("description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("metaTitle") {
		newMetaTitle := commercetools.LocalizedString(expandStringMap(d.Get("metaTitle").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetDescriptionAction{Description: &newMetaTitle})
	}

	if d.HasChange("metaDescription") {
		newMetaDescription := commercetools.LocalizedString(expandStringMap(d.Get("metaDescription").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetDescriptionAction{Description: &newMetaDescription})
	}

	if d.HasChange("metaKeywords") {
		newMetaKeywords := commercetools.LocalizedString(expandStringMap(d.Get("metaKeywords").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetDescriptionAction{Description: &newMetaKeywords})
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.CategoryUpdateWithID(context.Background(), input)
	if err != nil {
		if ctErr, ok := err.(commercetools.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return err
	}

	return resourceCategoryRead(d, m)
}

func resourceCategoryDelete(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)

	// Lock to prevent concurrent updates due to Version number conflicts
	ctMutexKV.Lock(d.Id())
	defer ctMutexKV.Unlock(d.Id())

	category, err := client.CategoryGetWithID(context.Background(), d.Id())
	if err != nil {
		return err
	}
	_, err = client.CategoryDeleteWithID(context.Background(), d.Id(), category.Version)
	if err != nil {
		return err
	}

	return nil
}
