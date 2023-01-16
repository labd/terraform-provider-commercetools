package commercetools

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

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

func migrateCategoryStateV0toV1(ctx context.Context, rawState map[string]any, meta any) (map[string]any, error) {
	for _, asset := range rawState["assets"].([]any) {
		sources := asset.(map[string]any)["sources"]
		for _, source := range sources.([]any) {
			transformToList(source.(map[string]any), "dimensions")
		}
	}
	return rawState, nil
}
