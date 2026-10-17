package integration_layer_field_visibility_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/integrationlayer"
	"github.com/stretchr/testify/assert"

	"github.com/labd/terraform-provider-commercetools/internal/acctest"
)

// projectKey is the Project of this test alone, with the Types and Product
// Types whose fields it hides.
const projectKey = "acc-cil-field-visibility"

// fields are the Custom Type and Product Type with the fields that can be
// hidden.
const fields = `
resource "commercetools_type" "order_extra" {
  key               = "order-extra"
  name              = { en = "Order extra" }
  resource_type_ids = ["order"]

  field {
    name     = "costPrice"
    label    = { en = "Cost price" }
    required = false
    type {
      name = "Money"
    }
  }
}

resource "commercetools_product_type" "chair" {
  key         = "chair"
  name        = "Chair"
  description = "Chair"

  attribute {
    name     = "supplierCode"
    label    = { en = "Supplier code" }
    required = false
    type {
      name = "text"
    }
  }
}
`

const twoFields = `
resource "commercetools_integration_layer_field_visibility" "visibility" {
  hidden_field {
    space      = "customField"
    type_key   = commercetools_type.order_extra.key
    field_name = "costPrice"
  }
  # Hidden before the attribute exists.
  hidden_field {
    space      = "attribute"
    type_key   = commercetools_product_type.chair.key
    field_name = "purchasePrice"
  }
}
`

const oneField = `
resource "commercetools_integration_layer_field_visibility" "visibility" {
  hidden_field {
    space      = "attribute"
    type_key   = commercetools_product_type.chair.key
    field_name = "supplierCode"
  }
}
`

func TestAccIntegrationLayerFieldVisibility(t *testing.T) {
	name := "commercetools_integration_layer_field_visibility.visibility"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: func(_ *terraform.State) error {
			for _, row := range visibility(t) {
				if row.Active || row.Reason != nil {
					return errors.New("fields are still hidden after destroy")
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: acctest.ProviderConfig(projectKey) + fields + twoFields,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "hidden_field.#", "2"),
					func(_ *terraform.State) error {
						rows := map[string]integrationlayer.FieldVisibilityRow{}
						for _, row := range visibility(t) {
							rows[row.Coordinate.TypeKey+"."+row.Coordinate.FieldName] = row
						}
						assert.True(t, rows["order-extra.costPrice"].Active)
						assert.False(t, rows["chair.supplierCode"].Active)
						assert.Equal(t, "noSuchField", *rows["chair.purchasePrice"].Reason)
						return nil
					},
				),
			},
			{
				Config: acctest.ProviderConfig(projectKey) + fields + oneField,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(name, "hidden_field.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(name, "hidden_field.*", map[string]string{
						"space": "attribute", "type_key": "chair", "field_name": "supplierCode",
					}),
				),
			},
			{
				Config:            acctest.ProviderConfig(projectKey) + fields + oneField,
				ResourceName:      name,
				ImportState:       true,
				ImportStateId:     projectKey,
				ImportStateVerify: true,
			},
		},
	})
}

func visibility(t *testing.T) []integrationlayer.FieldVisibilityRow {
	client, err := acctest.GetIntegrationLayerClient(projectKey)
	if err != nil {
		t.Fatal(err)
	}
	view, err := client.Main().Schema().Visibility().Get().Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return view.Fields
}

// TestIntegrationLayerFieldVisibilityPublishPending covers a write whose
// schema republish fails, which the mock cannot do: the rules are stored,
// the apply fails, and the next apply of the same configuration retries.
func TestIntegrationLayerFieldVisibilityPublishPending(t *testing.T) {
	var hidden []integrationlayer.SourceCoordinate
	failNext := false
	mux := http.NewServeMux()
	write := func(w http.ResponseWriter, body any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(body)
	}
	mux.HandleFunc("POST /oauth/token", func(w http.ResponseWriter, _ *http.Request) {
		write(w, map[string]any{"access_token": "token", "token_type": "Bearer", "expires_in": 3600})
	})
	mux.HandleFunc("GET /test-project/main/schema/visibility", func(w http.ResponseWriter, _ *http.Request) {
		rows := []integrationlayer.FieldVisibilityRow{}
		for _, c := range hidden {
			rows = append(rows, integrationlayer.FieldVisibilityRow{Coordinate: c, Active: true})
		}
		write(w, integrationlayer.FieldVisibilityView{Fields: rows, Version: 1})
	})
	mux.HandleFunc("PUT /test-project/main/schema/visibility", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&hidden)
		pending := failNext
		failNext = false
		write(w, integrationlayer.FieldVisibilityReplaceResult{HiddenFields: hidden, Version: 2, PublishPending: pending})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	providerConfig := fmt.Sprintf(`
provider "commercetools" {
  client_id             = "client"
  client_secret         = "secret"
  project_key           = "test-project"
  scopes                = "manage_project:test-project"
  api_url               = %[1]q
  token_url             = %[1]q
  integration_layer_url = %[1]q
}
`, server.URL)
	config := providerConfig + `
resource "commercetools_integration_layer_field_visibility" "visibility" {
  hidden_field {
    space      = "attribute"
    type_key   = "chair"
    field_name = "supplierCode"
  }
}
`

	resource.UnitTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.FrameworkProviderFactories(),
		Steps: []resource.TestStep{
			{
				PreConfig:   func() { failNext = true },
				Config:      config,
				ExpectError: regexp.MustCompile(`republishing the schema failed`),
			},
			{
				// The same configuration again retries the republish.
				Config: config,
				Check:  resource.TestCheckResourceAttr("commercetools_integration_layer_field_visibility.visibility", "hidden_field.#", "1"),
			},
		},
	})
}
