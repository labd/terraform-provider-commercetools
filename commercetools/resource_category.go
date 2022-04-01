package commercetools

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/labd/commercetools-go-sdk/platform"
)

func resourceCategory() *schema.Resource {
	return &schema.Resource{
		Description: "Categories allow you to organize products into hierarchical structures.\n\n" +
			"Also see the [Categories HTTP API documentation](https://docs.commercetools.com/api/projects/categories).",
		CreateContext: resourceCategoryCreate,
		ReadContext:   resourceCategoryRead,
		UpdateContext: resourceCategoryUpdate,
		DeleteContext: resourceCategoryDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceCategoryResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: migrateCategoryStateV0toV1,
				Version: 0,
			},
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Category-specific unique identifier. Must be unique across a project",
			},
			"name": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Required:         true,
				ForceNew:         true,
			},
			"description": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"slug": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Required:         true,
				Description:      "Human readable identifiers, needs to be unique",
			},
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A category that is the parent of this category in the category tree",
			},
			"order_hint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An attribute as base for a custom category order in one level, filled with random value when left empty",
			},
			"meta_title": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"meta_description": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"meta_keywords": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"assets": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Can be used to store images, icons or movies related to this category",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional User-defined identifier for the asset. Asset keys are unique inside their container (in this case the category)",
						},
						"name": {
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Required:         true,
						},
						"description": {
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Optional:         true,
						},
						"sources": {
							Type:        schema.TypeList,
							Optional:    true,
							MinItems:    1,
							Description: "Array of AssetSource, Has at least one entry",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"key": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Unique identifier, must be unique within the Asset",
									},
									"dimensions": {
										Type:     schema.TypeList,
										MaxItems: 1,
										Optional: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"w": {
													Type:        schema.TypeInt,
													Required:    true,
													Description: "The width of the asset source",
												},
												"h": {
													Type:        schema.TypeInt,
													Required:    true,
													Description: "The height of the asset source",
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

func resourceCategoryCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	var category *platform.Category

	name := unmarshallLocalizedString(d.Get("name"))
	slug := unmarshallLocalizedString(d.Get("slug"))

	draft := platform.CategoryDraft{
		Key:       stringRef(d.Get("key")),
		Name:      name,
		Slug:      slug,
		OrderHint: stringRef(d.Get("order_hint")),
	}

	if d.Get("description") != nil {
		desc := unmarshallLocalizedString(d.Get("description"))
		draft.Description = &desc
	}

	if d.Get("meta_title") != nil {
		metaTitle := unmarshallLocalizedString(d.Get("meta_title"))
		draft.MetaTitle = &metaTitle
	}

	if d.Get("meta_description") != nil {
		metaDescription := unmarshallLocalizedString(d.Get("meta_description"))
		draft.MetaDescription = &metaDescription
	}

	if d.Get("meta_keywords") != nil {
		metaKeywords := unmarshallLocalizedString(d.Get("meta_keywords"))
		draft.MetaKeywords = &metaKeywords
	}

	if d.Get("parent").(string) != "" {
		parent := platform.CategoryResourceIdentifier{}
		parent.ID = stringRef(d.Get("parent"))
		draft.Parent = &parent
	}

	if len(d.Get("assets").([]interface{})) != 0 {
		assets := unmarshallCategoryAssets(d)
		draft.Assets = assets
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error

		category, err = client.Categories().Post(draft).Execute(ctx)

		if err != nil {
			return handleCommercetoolsError(err)
		}
		return nil
	})

	if err != nil {
		return diag.FromErr(err)
	}

	if category == nil {
		return diag.Errorf("No  category created?")
	}

	d.SetId(category.ID)
	_ = d.Set("version", category.Version)

	return resourceCategoryRead(ctx, d, m)
}

func resourceCategoryRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	log.Printf("[DEBUG] Reading category from commercetools, with category id: %s", d.Id())
	client := getClient(m)

	category, err := client.Categories().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			if ctErr.StatusCode == 404 {
				d.SetId("")
				return nil
			}
		}
		return diag.FromErr(err)
	}

	if category == nil {
		log.Print("[DEBUG] No category found")
		d.SetId("")
	} else {
		log.Print("[DEBUG] Found following category:")
		log.Print(stringFormatObject(category))

		d.Set("version", category.Version)
		d.Set("key", category.Key)
		d.Set("name", category.Name)
		if category.Parent != nil {
			d.Set("parent", category.Parent.ID)
		} else {
			d.Set("parent", "")
		}
		d.Set("order_hint", category.OrderHint)
		if category.Description != nil {
			d.Set("description", *category.Description)
		}
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
			d.Set("assets", marshallCategoryAssets(category.Assets))
		}
	}
	return nil
}

func resourceCategoryUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)
	category, err := client.Categories().WithId(d.Id()).Get().Execute(ctx)

	if err != nil {
		return diag.FromErr(err)
	}

	input := platform.CategoryUpdate{
		Version: category.Version,
		Actions: []platform.CategoryUpdateAction{},
	}

	if d.HasChange("name") {
		newName := unmarshallLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.CategoryChangeNameAction{Name: newName})
	}

	if d.HasChange("slug") {
		newSlug := unmarshallLocalizedString(d.Get("slug"))
		input.Actions = append(
			input.Actions,
			&platform.CategoryChangeSlugAction{Slug: newSlug})
	}

	if d.HasChange("key") {
		newKey := d.Get("key").(string)
		input.Actions = append(
			input.Actions,
			&platform.CategorySetKeyAction{Key: &newKey})
	}

	if d.HasChange("order_hint") {
		newVal := d.Get("order_hint").(string)
		input.Actions = append(
			input.Actions,
			&platform.CategoryChangeOrderHintAction{OrderHint: newVal})
	}

	if d.HasChange("description") {
		newDescription := unmarshallLocalizedString(d.Get("description"))
		input.Actions = append(
			input.Actions,
			&platform.CategorySetDescriptionAction{Description: &newDescription})
	}

	if d.HasChange("parent") {
		newParentCategoryId := d.Get("parent").(string)
		parentId := platform.CategoryResourceIdentifier{ID: &newParentCategoryId}
		input.Actions = append(
			input.Actions,
			&platform.CategoryChangeParentAction{Parent: parentId})
	}

	if d.HasChange("meta_title") {
		newMetaTitle := unmarshallLocalizedString(d.Get("meta_title"))
		input.Actions = append(
			input.Actions,
			&platform.CategorySetMetaTitleAction{MetaTitle: &newMetaTitle})
	}

	if d.HasChange("meta_description") {
		newMetaDescription := unmarshallLocalizedString(d.Get("meta_description"))
		input.Actions = append(
			input.Actions,
			&platform.CategorySetMetaDescriptionAction{MetaDescription: &newMetaDescription})
	}

	if d.HasChange("meta_keywords") {
		newMetaKeywords := unmarshallLocalizedString(d.Get("meta_keywords"))
		input.Actions = append(
			input.Actions,
			&platform.CategorySetMetaKeywordsAction{MetaKeywords: &newMetaKeywords})
	}

	// TODO: This is far from complete. See
	// https://github.com/labd/terraform-provider-commercetools/issues/205
	if d.HasChange("assets") {
		assets := unmarshallCategoryAssets(d)
		for _, asset := range assets {
			input.Actions = append(
				input.Actions,
				&platform.CategoryChangeAssetNameAction{Name: asset.Name, AssetKey: asset.Key},
				&platform.CategorySetAssetDescriptionAction{Description: asset.Description, AssetKey: asset.Key},
				&platform.CategorySetAssetSourcesAction{Sources: asset.Sources, AssetKey: asset.Key},
			)
			if len(asset.Tags) > 0 {
				input.Actions = append(
					input.Actions,
					&platform.CategorySetAssetTagsAction{Tags: asset.Tags, AssetKey: asset.Key},
				)
			}
		}
	}

	log.Printf(
		"[DEBUG] Will perform update operation with the following actions:\n%s",
		stringFormatActions(input.Actions))

	_, err = client.Categories().WithId(d.Id()).Post(input).Execute(ctx)
	if err != nil {
		if ctErr, ok := err.(platform.ErrorResponse); ok {
			log.Printf("[DEBUG] %v: %v", ctErr, stringFormatErrorExtras(ctErr))
		}
		return diag.FromErr(err)
	}

	return resourceCategoryRead(ctx, d, m)
}

func resourceCategoryDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := getClient(m)

	version := d.Get("version").(int)
	_, err := client.Categories().WithId(d.Id()).Delete().Version(version).Execute(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func marshallCategoryAssets(assets []platform.Asset) []map[string]interface{} {

	result := make([]map[string]interface{}, len(assets))

	for i := range assets {
		asset := assets[i]

		result[i] = make(map[string]interface{})
		result[i]["name"] = asset.Name
		result[i]["key"] = asset.Key
		result[i]["sources"] = marshallCategoryAssetSources(asset.Sources)
		result[i]["tags"] = asset.Tags

		if asset.Description != nil {
			result[i]["description"] = *asset.Description
		} else {
			result[i]["description"] = nil
		}
	}

	return result
}

func unmarshallCategoryAssets(d *schema.ResourceData) []platform.AssetDraft {
	input := d.Get("assets").([]interface{})
	var result []platform.AssetDraft

	for _, raw := range input {
		i := raw.(map[string]interface{})

		name := unmarshallLocalizedString(i["name"])
		description := unmarshallLocalizedString(i["description"])
		sources := unmarshallCategoryAssetSources(i)
		tags := expandStringArray(i["tags"].([]interface{}))

		key := i["key"].(string)
		result = append(result, platform.AssetDraft{
			Key:         &key,
			Name:        name,
			Description: &description,
			Sources:     sources,
			Tags:        tags,
		})
	}

	return result
}

func marshallCategoryAssetSources(sources []platform.AssetSource) []map[string]interface{} {
	result := make([]map[string]interface{}, len(sources))

	for i := range sources {
		source := sources[i]

		result[i] = make(map[string]interface{})
		result[i]["key"] = source.Key
		result[i]["uri"] = source.Uri
		result[i]["content_type"] = source.ContentType

		if source.Dimensions != nil {
			result[i]["dimensions"] = []map[string]interface{}{
				{
					"h": source.Dimensions.H,
					"w": source.Dimensions.W,
				},
			}
		}
	}
	return result
}

func unmarshallCategoryAssetSources(i map[string]interface{}) []platform.AssetSource {
	var sources []platform.AssetSource
	for _, item := range i["sources"].([]interface{}) {
		s := item.(map[string]interface{})
		key := s["key"].(string)
		contentType := s["content_type"].(string)

		source := platform.AssetSource{
			Uri:         s["uri"].(string),
			Key:         &key,
			ContentType: &contentType,
		}

		if _, ok := s["dimensions"]; ok {
			source.Dimensions = unmarshallCategoryAssetSourceDimensions(s)
		}

		sources = append(sources, source)
	}
	return sources
}

func unmarshallCategoryAssetSourceDimensions(s map[string]interface{}) *platform.AssetDimensions {
	data, err := elementFromSlice(s, "dimensions")
	if err != nil {
		return nil
	}

	if data["w"] == nil || data["h"] == nil {
		return nil
	}

	return &platform.AssetDimensions{
		W: data["w"].(int),
		H: data["h"].(int),
	}
}

func resourceCategoryResourceV0() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Category-specific unique identifier. Must be unique across a project",
			},
			"name": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Required:         true,
				ForceNew:         true,
			},
			"description": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"slug": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Required:         true,
				Description:      "Human readable identifiers, needs to be unique",
			},
			"parent": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A category that is the parent of this category in the category tree",
			},
			"order_hint": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An attribute as base for a custom category order in one level, filled with random value when left empty",
			},
			"meta_title": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"meta_description": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"meta_keywords": {
				Type:             TypeLocalizedString,
				ValidateDiagFunc: validateLocalizedStringKey,
				Optional:         true,
			},
			"assets": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Can be used to store images, icons or movies related to this category",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Optional User-defined identifier for the asset. Asset keys are unique inside their container (in this case the category)",
						},
						"name": {
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Required:         true,
						},
						"description": {
							Type:             TypeLocalizedString,
							ValidateDiagFunc: validateLocalizedStringKey,
							Optional:         true,
						},
						"sources": {
							Type:        schema.TypeList,
							Optional:    true,
							MinItems:    1,
							Description: "Array of AssetSource, Has at least one entry",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"uri": {
										Type:     schema.TypeString,
										Required: true,
									},
									"key": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Unique identifier, must be unique within the Asset",
									},
									"dimensions": {
										Type:     schema.TypeMap,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeInt,
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

func migrateCategoryStateV0toV1(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
	for _, asset := range rawState["assets"].([]interface{}) {
		sources := asset.(map[string]interface{})["sources"]
		for _, source := range sources.([]interface{}) {
			transformToList(source.(map[string]interface{}), "dimensions")
		}
	}
	return rawState, nil
}
