package commercetools

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/stretchr/testify/assert"
)

func TestCustomApplicationNavbarMenuConversion(t *testing.T) {
	resourceDataMap := []map[string]interface{}{
		{
			"uri_path":    "state-machines",
			"icon":        "RocketIcon",
			"permissions": []string{"ViewDeveloperSettings"},
			"label_all_locales": []map[string]interface{}{
				{
					"locale": "en",
					"value":  "State machines",
				},
				{
					"locale": "de",
					"value":  "Zustandsmachinen",
				},
			},
			"submenu": []map[string]interface{}{
				{
					"uri_path":    "state-machines/new",
					"permissions": []string{"ManageDeveloperSettings"},
					"label_all_locales": []map[string]interface{}{
						{
							"locale": "en",
							"value":  "Add state machine",
						},
						{
							"locale": "de",
							"value":  "Zustandsmachine hinzufügen",
						},
					},
				},
			},
		},
	}

	d := schema.TestResourceDataRaw(t, resourceCustomApplication().Schema, nil)
	err := d.Set("navbar_menu", resourceDataMap)
	if err != nil {
		t.Error("Failed to set the navbar_menu value to the resource data")
	}
	doc := resourceCustomApplicationFormToDocNavbarMenu(d.Get("navbar_menu").([]interface{}))
	navbarMenu := NavbarMenu{}
	jsonString, err := json.Marshal(doc)
	if err != nil {
		t.Error("Failed to convert navbar menu from map to struct")
	}
	json.Unmarshal(jsonString, &navbarMenu)
	form := resourceCustomApplicationDocToFormNavbarMenu(navbarMenu)

	assert.Equal(t, len(form), 1, "There should only be 1 navbarMenu")
	assert.Equal(t, doc["key"], doc["uriPath"], "The menu key is computed from the uriPath")
	assert.Equal(t, form[0]["uri_path"], doc["uriPath"], "Converted uriPath should match")

	docSubmenu := doc["submenu"].([]map[string]interface{})
	formSubmenu := form[0]["submenu"].([]map[string]interface{})
	assert.Equal(t, len(formSubmenu), 1, "There is 1 item in the submenu list")
	assert.Equal(t, docSubmenu[0]["key"], "state-machines-new", "The menu key is computed from the uriPath")
}
