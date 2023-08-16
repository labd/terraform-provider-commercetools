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
				Description: "Category-specific unique identifier. Must be unique across a project",
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
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "",
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
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
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
			"custom": CustomFieldSchema(),
			"version": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceCategoryCreate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	name := expandLocalizedString(d.Get("name"))
	slug := expandLocalizedString(d.Get("slug"))
	key := stringRef(d.Get("key"))

	custom, err := CreateCustomFieldDraft(ctx, client, d)
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	draft := platform.CategoryDraft{
		Name:      name,
		Slug:      slug,
		OrderHint: stringRef(d.Get("order_hint")),
		Custom:    custom,
	}

	if *key != "" {
		draft.Key = key
	}

	if d.Get("description") != nil {
		desc := expandLocalizedString(d.Get("description"))
		draft.Description = &desc
	}

	if d.Get("meta_title") != nil {
		metaTitle := expandLocalizedString(d.Get("meta_title"))
		draft.MetaTitle = &metaTitle
	}

	if d.Get("meta_description") != nil {
		metaDescription := expandLocalizedString(d.Get("meta_description"))
		draft.MetaDescription = &metaDescription
	}

	if d.Get("meta_keywords") != nil {
		metaKeywords := expandLocalizedString(d.Get("meta_keywords"))
		draft.MetaKeywords = &metaKeywords
	}

	if d.Get("parent").(string) != "" {
		parent := platform.CategoryResourceIdentifier{}
		parent.ID = stringRef(d.Get("parent"))
		draft.Parent = &parent
	}

	if len(d.Get("assets").([]any)) != 0 {
		assets := expandCategoryAssetDrafts(d)
		draft.Assets = assets
	}

	if d.Get("external_id").(string) != "" {
		draft.ExternalId = stringRef(d.Get("external_id"))
	}

	var category *platform.Category
	err = resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		var err error
		category, err = client.Categories().Post(draft).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})

	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(category.ID)
	_ = d.Set("version", category.Version)

	return resourceCategoryRead(ctx, d, m)
}

func resourceCategoryRead(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)
	category, err := client.Categories().WithId(d.Id()).Get().Execute(ctx)
	if err != nil {
		if utils.IsResourceNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	_ = d.Set("version", category.Version)
	_ = d.Set("key", category.Key)
	_ = d.Set("name", category.Name)
	if category.Parent != nil {
		_ = d.Set("parent", category.Parent.ID)
	} else {
		_ = d.Set("parent", "")
	}
	_ = d.Set("order_hint", category.OrderHint)
	_ = d.Set("external_id", category.ExternalId)
	if category.Description != nil {
		_ = d.Set("description", *category.Description)
	}
	if category.MetaTitle != nil {
		_ = d.Set("meta_title", *category.MetaTitle)
	}
	if category.MetaDescription != nil {
		_ = d.Set("meta_description", *category.MetaDescription)
	}
	if category.MetaKeywords != nil {
		_ = d.Set("meta_keywords", *category.MetaKeywords)
	}
	if category.Assets != nil {
		_ = d.Set("assets", flattenCategoryAssets(category.Assets))
	}
	_ = d.Set("custom", flattenCustomFields(category.Custom))
	return nil
}

func resourceCategoryUpdate(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	input := platform.CategoryUpdate{
		Version: d.Get("version").(int),
		Actions: []platform.CategoryUpdateAction{},
	}

	if d.HasChange("name") {
		newName := expandLocalizedString(d.Get("name"))
		input.Actions = append(
			input.Actions,
			&platform.CategoryChangeNameAction{Name: newName})
	}

	if d.HasChange("slug") {
		newSlug := expandLocalizedString(d.Get("slug"))
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

	if d.HasChange("external_id") {
		newExternalID := d.Get("external_id").(string)
		input.Actions = append(
			input.Actions,
			&platform.CategorySetExternalIdAction{ExternalId: &newExternalID})
	}

	if d.HasChange("description") {
		newDescription := expandLocalizedString(d.Get("description"))
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
		newMetaTitle := expandLocalizedString(d.Get("meta_title"))
		input.Actions = append(
			input.Actions,
			&platform.CategorySetMetaTitleAction{MetaTitle: &newMetaTitle})
	}

	if d.HasChange("meta_description") {
		newMetaDescription := expandLocalizedString(d.Get("meta_description"))
		input.Actions = append(
			input.Actions,
			&platform.CategorySetMetaDescriptionAction{MetaDescription: &newMetaDescription})
	}

	if d.HasChange("meta_keywords") {
		newMetaKeywords := expandLocalizedString(d.Get("meta_keywords"))
		input.Actions = append(
			input.Actions,
			&platform.CategorySetMetaKeywordsAction{MetaKeywords: &newMetaKeywords})
	}

	if d.HasChange("assets") {
		// Instead of having to perform operations on lists of assets we remove all assets on change and recreate the new
		//list
		oldState, newState := d.GetChange("assets")

		oldAssets, ok := oldState.([]interface{})
		if !ok {
			return diag.Errorf("old asset state is not a list")
		}

		for _, assetData := range oldAssets {
			asset, ok := assetData.(map[string]interface{})
			if !ok {
				return diag.Errorf("asset is not in format map[string]interface{}")
			}
			id := asset["id"].(string)
			input.Actions = append(input.Actions, &platform.CategoryRemoveAssetAction{AssetId: &id})
		}

		newAssets, ok := newState.([]interface{})
		if !ok {
			return diag.Errorf("new asset state is not a list")
		}

		for _, assetData := range newAssets {
			if !ok {
				return diag.Errorf("asset is not in format map[string]interface{}")
			}
			input.Actions = append(input.Actions, &platform.CategoryAddAssetAction{
				Asset: *expandCategoryAssetDraft(assetData),
			})
		}
	}

	if d.HasChange("custom") {
		actions, err := CustomFieldUpdateActions[platform.CategorySetCustomTypeAction, platform.CategorySetCustomFieldAction](ctx, client, d)
		if err != nil {
			return diag.FromErr(err)
		}
		for i := range actions {
			input.Actions = append(input.Actions, actions[i].(platform.CategoryUpdateAction))
		}
	}

	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.Categories().WithId(d.Id()).Post(input).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		// Workaround invalid state to be written, see
		// https://github.com/hashicorp/terraform-plugin-sdk/issues/476
		d.Partial(true)
		return diag.FromErr(err)
	}

	return resourceCategoryRead(ctx, d, m)
}

func resourceCategoryDelete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics {
	client := getClient(m)

	version := d.Get("version").(int)
	err := resource.RetryContext(ctx, 1*time.Minute, func() *resource.RetryError {
		_, err := client.Categories().WithId(d.Id()).Delete().Version(version).Execute(ctx)
		return utils.ProcessRemoteError(err)
	})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func flattenCategoryAssets(assets []platform.Asset) []map[string]any {
	result := make([]map[string]any, len(assets))

	for i := range assets {
		asset := assets[i]

		result[i] = make(map[string]any)
		if asset.ID != "" {
			result[i]["id"] = asset.ID
		}
		result[i]["name"] = asset.Name
		result[i]["key"] = asset.Key
		result[i]["sources"] = flattenCategoryAssetSources(asset.Sources)
		result[i]["tags"] = asset.Tags

		if asset.Description != nil {
			result[i]["description"] = *asset.Description
		} else {
			result[i]["description"] = nil
		}
	}

	return result
}

func expandCategoryAssetDraft(u any) *platform.AssetDraft {
	i := u.(map[string]any)
	name := expandLocalizedString(i["name"])
	description := expandLocalizedString(i["description"])
	sources := expandCategoryAssetSources(i)
	tags := expandStringArray(i["tags"].([]any))
	key := i["key"].(string)
	return &platform.AssetDraft{
		Key:         &key,
		Name:        name,
		Description: &description,
		Sources:     sources,
		Tags:        tags,
	}
}

func expandCategoryAssetDrafts(d *schema.ResourceData) []platform.AssetDraft {
	input := d.Get("assets").([]any)
	var result []platform.AssetDraft
	for _, raw := range input {
		result = append(result, *expandCategoryAssetDraft(raw))
	}
	return result
}

func expandCategoryAssets(d *schema.ResourceData) []platform.Asset {
	input := d.Get("assets").([]any)
	var result []platform.Asset
	for _, raw := range input {
		draft := expandCategoryAssetDraft(raw)
		i := raw.(map[string]any)
		id := i["id"].(string)
		asset := platform.Asset{
			ID:          id,
			Key:         draft.Key,
			Name:        draft.Name,
			Description: draft.Description,
			Sources:     draft.Sources,
			Tags:        draft.Tags,
		}
		result = append(result, asset)
	}
	return result
}

func flattenCategoryAssetSources(sources []platform.AssetSource) []map[string]any {
	result := make([]map[string]any, len(sources))

	for i := range sources {
		source := sources[i]

		result[i] = make(map[string]any)
		result[i]["key"] = source.Key
		result[i]["uri"] = source.Uri
		result[i]["content_type"] = source.ContentType

		if source.Dimensions != nil {
			result[i]["dimensions"] = []map[string]any{
				{
					"h": source.Dimensions.H,
					"w": source.Dimensions.W,
				},
			}
		}
	}
	return result
}

func expandCategoryAssetSources(i map[string]any) []platform.AssetSource {
	var sources []platform.AssetSource
	for _, item := range i["sources"].([]any) {
		s := item.(map[string]any)
		key := s["key"].(string)
		contentType := s["content_type"].(string)

		source := platform.AssetSource{
			Uri:         s["uri"].(string),
			Key:         &key,
			ContentType: &contentType,
		}

		if _, ok := s["dimensions"]; ok {
			source.Dimensions = expandCategoryAssetSourceDimensions(s)
		}

		sources = append(sources, source)
	}
	return sources
}

func expandCategoryAssetSourceDimensions(s map[string]any) *platform.AssetDimensions {
	data := elementFromSlice(s, "dimensions")
	if data == nil {
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
