package commercetools

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/machinebox/graphql"
)

func resourceCustomApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomApplicationCreate,
		Read:   resourceCustomApplicationRead,
		Update: resourceCustomApplicationUpdate,
		Delete: resourceCustomApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"is_active": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			// NOTE: currently the terraform SDK does not yet support nested object maps,
			// even though terraform 0.12 has support for that. As a workaround we can use
			// a `TypeList` with a `MaxItems` of 1
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/62
			// https://github.com/hashicorp/terraform-plugin-sdk/issues/155.
			"navbar_menu": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"uri_path": {
							Type:     schema.TypeString,
							Required: true,
						},
						"icon": {
							Type:     schema.TypeString,
							Required: true,
						},
						"permissions": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"label_all_locales": {
							Type:     schema.TypeList,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"locale": {
										Type:     schema.TypeString,
										Required: true,
									},
									"value": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"submenu": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"key": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"uri_path": {
										Type:     schema.TypeString,
										Required: true,
									},
									"permissions": {
										Type:     schema.TypeList,
										Optional: true,
										Elem: &schema.Schema{
											Type: schema.TypeString,
										},
									},
									"label_all_locales": {
										Type:     schema.TypeList,
										Required: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"locale": {
													Type:     schema.TypeString,
													Required: true,
												},
												"value": {
													Type:     schema.TypeString,
													Required: true,
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func resourceCustomApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	client := getGraphQLClient(meta)
	projectKey := getProjectKey(meta)

	req := graphql.NewRequest(`
		mutation CreateCustomApplication($draft: ApplicationExtensionDataInput!) {
			createProjectExtensionApplication(data: $draft) {
				id
				applications {
					id
				}
			}
		}
	`)
	draft := map[string]interface{}{
		"name":        d.Get("name"),
		"description": d.Get("description"),
		"url":         d.Get("url"),
		"navbarMenu":  resourceCustomApplicationFormToDocNavbarMenu(d.Get("navbar_menu").([]interface{})),
	}
	req.Var("draft", draft)
	req.Header.Set("X-Project-Key", projectKey)
	req.Header.Set("X-GraphQL-Target", "settings")

	var res GraphQLResponseProjectExtensionCreation
	err := client.Run(context.Background(), req, &res)
	if err != nil {
		return err
	}
	customApp := res.CreateProjectExtensionApplication.Applications[0]

	if d.Get("is_active").(bool) {
		reqToActivate := graphql.NewRequest(`
			mutation ActivateCustomApplicationAfterCreation($applicationId: ID!) {
				activateProjectExtensionApplication(applicationId: $applicationId) {
					id
					applications(where: { id: $applicationId }) {
						id
					}
				}
			}
		`)
		reqToActivate.Var("applicationId", customApp.ID)
		reqToActivate.Header.Set("X-Project-Key", projectKey)
		reqToActivate.Header.Set("X-GraphQL-Target", "settings")

		var resFromActivate GraphQLResponseProjectExtensionUpdate
		errFromActivate := client.Run(context.Background(), reqToActivate, &resFromActivate)
		if errFromActivate != nil {
			return errFromActivate
		}
	}

	d.SetId(customApp.ID)
	return resourceCustomApplicationRead(d, meta)
}

func resourceCustomApplicationRead(d *schema.ResourceData, meta interface{}) error {
	client := getGraphQLClient(meta)
	projectKey := getProjectKey(meta)

	req := graphql.NewRequest(`
		query FetchCustomApplicationById($applicationId: ID!) {
			projectExtension {
				id
				applications(where: { id: $applicationId }) {
					id
					createdAt
					updatedAt
					isActive
					name
					description
					url
					navbarMenu {
						uriPath
						icon
						labelAllLocales {
							locale
							value
						}
						permissions
						submenu {
							uriPath
							labelAllLocales {
								locale
								value
							}
							permissions
						}
					}
				}
			}
		}
	`)
	req.Var("applicationId", d.Id())
	req.Header.Set("X-Project-Key", projectKey)
	req.Header.Set("X-GraphQL-Target", "settings")

	var res GraphQLResponseProjectExtension
	err := client.Run(context.Background(), req, &res)

	if err != nil {
		d.SetId("")
		return err
	}

	if res.ProjectExtension == nil {
		log.Print("[DEBUG] No project extension found")
		d.SetId("")
		return nil
	}

	if len(res.ProjectExtension.Applications) == 0 {
		log.Print("[DEBUG] No custom application found")
		d.SetId("")
		return nil
	}

	customApp := res.ProjectExtension.Applications[0]

	// Assign values from response
	d.SetId(customApp.ID)
	d.Set("created_at", customApp.CreatedAt)
	d.Set("updated_at", customApp.UpdatedAt)
	d.Set("is_active", customApp.IsActive)
	d.Set("name", customApp.Name)
	if customApp.Description != nil {
		d.Set("description", customApp.Description)
	}
	d.Set("url", customApp.URL)
	d.Set("navbar_menu", resourceCustomApplicationDocToFormNavbarMenu(customApp.NavbarMenu))
	return nil
}

func resourceCustomApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := getGraphQLClient(meta)
	projectKey := getProjectKey(meta)

	req := graphql.NewRequest(`
		mutation UpdateCustomApplication(
			$applicationId: ID!
			$draft: ApplicationExtensionDataInput!
			$shouldActivate: Boolean!
		) {
			updateProjectExtensionApplication(
				applicationId: $applicationId
				data: $draft
			) {
				id
				applications(where: { id: $applicationId }) {
					id
				}
			}
			activateProjectExtensionApplication(applicationId: $applicationId) @include(if: $shouldActivate) {
				id
				applications(where: { id: $applicationId }) {
					id
				}
			}
			deactivateProjectExtensionApplication(applicationId: $applicationId) @skip(if: $shouldActivate) {
				id
				applications(where: { id: $applicationId }) {
					id
				}
			}
		}
	`)
	draft := map[string]interface{}{
		"name":        d.Get("name"),
		"description": d.Get("description"),
		"url":         d.Get("url"),
		"navbarMenu":  resourceCustomApplicationFormToDocNavbarMenu(d.Get("navbar_menu").([]interface{})),
	}
	req.Var("applicationId", d.Id())
	req.Var("draft", draft)
	req.Var("shouldActivate", d.Get("is_active"))
	req.Header.Set("X-Project-Key", projectKey)
	req.Header.Set("X-GraphQL-Target", "settings")

	var res GraphQLResponseProjectExtensionUpdate
	err := client.Run(context.Background(), req, &res)

	if err != nil {
		return err
	}

	return resourceCustomApplicationRead(d, meta)
}

func resourceCustomApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	client := getGraphQLClient(meta)
	projectKey := getProjectKey(meta)

	req := graphql.NewRequest(`
		mutation DeleteCustomApplication($applicationId: ID!) {
			deleteProjectExtensionApplication(applicationId: $applicationId) {
				id
				applications {
					id
				}
			}
		}
	`)
	req.Var("applicationId", d.Id())
	req.Header.Set("X-Project-Key", projectKey)
	req.Header.Set("X-GraphQL-Target", "settings")

	var res GraphQLResponseProjectExtensionDeletion
	err := client.Run(context.Background(), req, &res)

	if err != nil {
		return err
	}

	// NOTE: `d.SetId("")`` is automatically called assuming delete returns no errors
	// https://www.terraform.io/docs/extend/writing-custom-providers.html#implementing-destroy
	return nil
}

/* Utility functions */

func resourceCustomApplicationFormToDocNavbarMenu(formValues []interface{}) map[string]interface{} {
	// There can only be one `navbarMenu`.
	value := (formValues[0]).(map[string]interface{})
	result := map[string]interface{}{
		"key":             slugify(value["uri_path"].(string)),
		"uriPath":         value["uri_path"],
		"icon":            value["icon"],
		"labelAllLocales": value["label_all_locales"],
	}
	// Set optional values
	if value["permissions"] == nil {
		result["permissions"] = make([]string, 0)
	} else {
		result["permissions"] = value["permissions"]
	}
	if value["submenu"] == nil {
		result["submenu"] = make([]map[string]interface{}, 0)
	} else {
		result["submenu"] = resourceCustomApplicationFormToDocNavbarSubmenu(value["submenu"].([]interface{}))
	}
	return result
}

func resourceCustomApplicationFormToDocNavbarSubmenu(formValues []interface{}) []map[string]interface{} {
	var result []map[string]interface{}
	for _, raw := range formValues {
		value := raw.(map[string]interface{})
		doc := map[string]interface{}{
			"key":             slugify(value["uri_path"].(string)),
			"uriPath":         value["uri_path"],
			"labelAllLocales": value["label_all_locales"],
		}
		// Set optional values
		if value["permissions"] == nil {
			doc["permissions"] = make([]string, 0)
		} else {
			doc["permissions"] = value["permissions"]
		}
		result = append(result, doc)
	}
	return result
}

// Map all the nested fields of the navbarMenu config to the corresponding terraform fields.
func resourceCustomApplicationDocToFormNavbarMenu(navbarMenu NavbarMenu) []map[string]interface{} {
	result := map[string]interface{}{
		"uri_path":          navbarMenu.URIPath,
		"icon":              navbarMenu.Icon,
		"permissions":       navbarMenu.Permissions,
		"label_all_locales": navbarMenu.LabelAllLocales,
	}
	submenu := make([]map[string]interface{}, len(navbarMenu.Submenu))
	for i, menu := range navbarMenu.Submenu {
		elem := map[string]interface{}{
			"uri_path":          menu.URIPath,
			"permissions":       menu.Permissions,
			"label_all_locales": menu.LabelAllLocales,
		}
		submenu[i] = elem
	}
	result["submenu"] = submenu
	formValues := []map[string]interface{}{result}
	return formValues
}
