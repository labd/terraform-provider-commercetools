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
				ForceNew: true,
			},
			"name": {
				Type:     TypeLocalizedString,
				Required: true,
				ForceNew: true,
			},
			"description": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"slug": {
				Type:     TypeLocalizedString,
				Required: true,
			},
			"parent": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"order_hint": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"meta_title": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"meta_description": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"meta_keywords": {
				Type:     TypeLocalizedString,
				Optional: true,
			},
			"assets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"name": {
							Type:     TypeLocalizedString,
							Required: true,
						},
						"description": {
							Type:     TypeLocalizedString,
							Optional: true,
						},
						"sources": {
							Type:     schema.TypeList,
							Optional: true,
							MinItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"key": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"dimensions": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"w": {
													Type:     schema.TypeInt,
													Required: true,
												},
												"h": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"content_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
								},
							},
						},
						"tags": {
							Type:     schema.TypeList,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
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

func resourceCategoryCreate(d *schema.ResourceData, m interface{}) error {
	client := getClient(m)
	var category *commercetools.Category

	name := commercetools.LocalizedString(expandStringMap(d.Get("name").(map[string]interface{})))
	desc := commercetools.LocalizedString(expandStringMap(d.Get("description").(map[string]interface{})))
	slug := commercetools.LocalizedString(expandStringMap(d.Get("slug").(map[string]interface{})))
	metaTitle := commercetools.LocalizedString(expandStringMap(d.Get("meta_title").(map[string]interface{})))
	metaDescription := commercetools.LocalizedString(expandStringMap(d.Get("meta_description").(map[string]interface{})))
	metaKeywords := commercetools.LocalizedString(expandStringMap(d.Get("meta_keywords").(map[string]interface{})))

	draft := &commercetools.CategoryDraft{
		Key:             d.Get("key").(string),
		Name:            &name,
		Description:     &desc,
		Slug:            &slug,
		OrderHint:       d.Get("order_hint").(string),
		MetaTitle:       &metaTitle,
		MetaDescription: &metaDescription,
		MetaKeywords:    &metaKeywords,
	}

	if d.Get("parent").(string) != "" {
		parent := commercetools.CategoryResourceIdentifier{}
		parent.ID = d.Get("parent").(string)
		draft.Parent = &parent
	}

	if len(d.Get("assets").([]interface{})) != 0 {
		assets := resourceCategoryGetAssets(d)
		draft.Assets = assets
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
	_ = d.Set("version", category.Version)

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
		d.Set("parent", category.Parent)
		d.Set("order_hint", category.OrderHint)
		if category.MetaTitle != nil {
			d.Set("meta_title", *category.MetaTitle)
		}
		if category.MetaDescription != nil {
			d.Set("meta_description", *category.MetaDescription)
		}
		if category.MetaKeywords != nil {
			d.Set("meta_keywords", *category.MetaKeywords)
		}
		if category.Assets != nil {
			d.Set("assets", category.Assets)
		}
	}
	return nil
}

func resourceCategoryUpdate(d *schema.ResourceData, m interface{}) error {
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

	if d.HasChange("order_hint") {
		newVal := d.Get("order_hint").(string)
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

	if d.HasChange("parent") {
		newParentCategoryId := d.Get("parent").(string)
		parentId := commercetools.CategoryResourceIdentifier{ID: newParentCategoryId}
		input.Actions = append(
			input.Actions,
			&commercetools.CategoryChangeParentAction{Parent: &parentId})
	}

	if d.HasChange("meta_title") {
		newMetaTitle := commercetools.LocalizedString(expandStringMap(d.Get("meta_title").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetMetaTitleAction{MetaTitle: &newMetaTitle})
	}

	if d.HasChange("meta_description") {
		newMetaDescription := commercetools.LocalizedString(expandStringMap(d.Get("meta_description").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetMetaDescriptionAction{MetaDescription: &newMetaDescription})
	}

	if d.HasChange("meta_keywords") {
		newMetaKeywords := commercetools.LocalizedString(expandStringMap(d.Get("meta_keywords").(map[string]interface{})))
		input.Actions = append(
			input.Actions,
			&commercetools.CategorySetMetaKeywordsAction{MetaKeywords: &newMetaKeywords})
	}

	if d.HasChange("assets") {
		assets := resourceCategoryGetAssets(d)
		for _, asset := range assets {
			input.Actions = append(
				input.Actions,
				&commercetools.CategoryChangeAssetNameAction{Name: asset.Name, AssetKey: asset.Key},
				&commercetools.CategorySetAssetDescriptionAction{Description: asset.Description, AssetKey: asset.Key},
				&commercetools.CategorySetAssetSourcesAction{Sources: asset.Sources, AssetKey: asset.Key},
			)
			if len(asset.Tags) > 0 {
				input.Actions = append(
					input.Actions,
					&commercetools.CategorySetAssetTagsAction{Tags: asset.Tags, AssetKey: asset.Key},
				)
			}
		}
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

	version := d.Get("version").(int)
	_, err := client.CategoryDeleteWithID(context.Background(), d.Id(), version)
	if err != nil {
		return err
	}

	return nil
}

func resourceCategoryGetAssets(d *schema.ResourceData) []commercetools.AssetDraft {
	input := d.Get("assets").([]interface{})
	var result []commercetools.AssetDraft

	for _, raw := range input {
		i := raw.(map[string]interface{})

		name := commercetools.LocalizedString(expandStringMap(i["name"].(map[string]interface{})))
		description := commercetools.LocalizedString(expandStringMap(i["description"].(map[string]interface{})))
		sources := resourceCategoryGetAssetSources(i)

		result = append(result, commercetools.AssetDraft{
			Key:         i["key"].(string),
			Name:        &name,
			Description: &description,
			Sources:     sources,
		})
	}

	return result
}

func resourceCategoryGetAssetSources(i map[string]interface{}) []commercetools.AssetSource {
	var sources []commercetools.AssetSource
	for _, item := range i["sources"].([]interface{}) {
		s := item.(map[string]interface{})

		source := commercetools.AssetSource{
			URI:         s["uri"].(string),
			Key:         s["key"].(string),
			ContentType: s["content_type"].(string),
		}

		if _, ok := s["dimensions"]; ok {
			assetDimensions := resourceCategoryGetAssetSourceDimensions(s)
			source.Dimensions = &assetDimensions
		}

		sources = append(sources, source)
	}
	return sources
}

func resourceCategoryGetAssetSourceDimensions(s map[string]interface{}) commercetools.AssetDimensions {
	var dimensions commercetools.AssetDimensions
	for _, item := range s["dimensions"].(map[string]interface{}) {
		d := item.(map[string]interface{})

		dimensions = commercetools.AssetDimensions{
			W: d["w"].(float64),
			H: d["h"].(float64),
		}
	}
	return dimensions
}
