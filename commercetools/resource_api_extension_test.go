package commercetools

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/labd/commercetools-go-sdk/platform"
	"github.com/stretchr/testify/assert"
)

func TestAPIExtensionExpandExtensionDestination(t *testing.T) {
	rawDestination := map[string]any{
		"type":          "AWSLambda",
		"arn":           "arn:aws:lambda:eu-west-1:111111111:function:api_extensions",
		"access_key":    "ABCSDF123123123",
		"access_secret": "****abc/",
	}

	resourceDataMap := map[string]any{
		"id":             "2845b936-e407-4f29-957b-f8deb0fcba97",
		"version":        1,
		"createdAt":      "2018-12-03T16:13:03.969Z",
		"lastModifiedAt": "2018-12-04T09:06:59.491Z",
		"destination":    []any{rawDestination},
		"triggers": []any{
			map[string]any{
				"triggers": []any{"Create", "Update"},
			},
		},
		"timeout_in_ms": 1,
		"key":           "create-order",
	}

	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, resourceDataMap)
	destination, _ := expandExtensionDestination(d)
	lambdaDestination, ok := destination.(platform.AWSLambdaDestination)

	assert.True(t, ok)
	assert.Equal(t, lambdaDestination.Arn, "arn:aws:lambda:eu-west-1:111111111:function:api_extensions")
	assert.Equal(t, lambdaDestination.AccessKey, "ABCSDF123123123")
	assert.Equal(t, lambdaDestination.AccessSecret, "****abc/")
}

func TestAPIExtensionExpandExtensionDestinationAuthentication(t *testing.T) {
	var input = map[string]any{
		"authorization_header": "12345",
		"azure_authentication": "AzureKey",
	}

	auth, err := expandExtensionDestinationAuthentication(input)
	assert.Nil(t, auth)
	assert.NotNil(t, err)

	input = map[string]any{
		"authorization_header": "12345",
	}

	auth, err = expandExtensionDestinationAuthentication(input)
	httpAuth, ok := auth.(platform.AuthorizationHeaderAuthentication)
	assert.True(t, ok)
	assert.Equal(t, "12345", httpAuth.HeaderValue)
	assert.NotNil(t, auth)
	assert.Nil(t, err)
}

func TestExpandExtensionTriggers(t *testing.T) {
	resourceDataMap := map[string]any{
		"id":             "2845b936-e407-4f29-957b-f8deb0fcba97",
		"version":        1,
		"createdAt":      "2018-12-03T16:13:03.969Z",
		"lastModifiedAt": "2018-12-04T09:06:59.491Z",
		"trigger": []any{
			map[string]any{
				"resource_type_id": "cart",
				"actions":          []any{"Create", "Update"},
			},
		},
		"timeout_in_ms": 1,
		"key":           "create-order",
	}

	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, resourceDataMap)
	triggers := expandExtensionTriggers(d)

	assert.Len(t, triggers, 1)
	assert.Equal(t, triggers[0].ResourceTypeId, platform.ExtensionResourceTypeIdCart)
	assert.Len(t, triggers[0].Actions, 2)
}

func TestExpandExtensionAdditionalContext(t *testing.T) {
	resourceDataMap := map[string]any{
		"trigger": []any{
			map[string]any{
				"resource_type_id": "cart",
				"actions":          []any{"Create"},
			},
		},
		"additional_context": []any{
			map[string]any{
				"include_old_resource": true,
			},
		},
	}

	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, resourceDataMap)
	ac := expandExtensionAdditionalContext(d.Get("additional_context").([]any))
	assert.NotNil(t, ac)
	assert.NotNil(t, ac.IncludeOldResource)
	assert.True(t, *ac.IncludeOldResource)

	assert.Nil(t, expandExtensionAdditionalContext([]any{}))
}

func TestFlattenExtensionAdditionalContext(t *testing.T) {
	assert.Nil(t, flattenExtensionAdditionalContext(nil))

	result := flattenExtensionAdditionalContext(&platform.ExtensionAdditionalContext{
		IncludeOldResource: true,
	})
	assert.Equal(t, []map[string]any{{"include_old_resource": true}}, result)
}

func TestAccAPIExtension_basic(t *testing.T) {
	name := fmt.Sprintf("extension_%s", acctest.RandString(5))
	timeoutInMs := acctest.RandIntRange(200, 1800)
	identifier := "ext"
	resourceName := "commercetools_api_extension.ext"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAPIExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIExtensionGCFConfig(identifier, name, timeoutInMs),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists("ext"),
					resource.TestCheckResourceAttr(
						resourceName, "key", name),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_in_ms", strconv.FormatInt(int64(timeoutInMs), 10)),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.0", "Create"),
				),
			},
			{
				Config: testAccAPIExtensionConfig(identifier, name, timeoutInMs),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists("ext"),
					resource.TestCheckResourceAttr(
						resourceName, "key", name),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_in_ms", strconv.FormatInt(int64(timeoutInMs), 10)),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.0", "Create"),
				),
			},
			{
				Config: testAccAPIExtensionConfigRequiredOnly(identifier, name),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "key", name),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.0", "Create"),
				),
			},
			{
				Config: testAccAPIExtensionUpdate(identifier, name, timeoutInMs),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "key", name),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_in_ms", strconv.FormatInt(int64(timeoutInMs), 10)),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.0", "Create"),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.1", "Update"),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.condition", "name = \"Michael\""),
				),
			},
		},
	})
}

func testAccAPIExtensionGCFConfig(identifier, key string, timeoutInMs int) string {
	return hclTemplate(`
		resource "commercetools_api_extension" "{{ .identifier }}" {
			key = "{{ .key }}"
			timeout_in_ms = {{ .timeoutInMs }}

			destination {
				type                 = "GoogleCloudFunction"
				url                  = "https://example.com"
			}

			trigger {
				resource_type_id = "customer"
				actions = ["Create"]
			}
		}
	`, map[string]any{
		"identifier":  identifier,
		"key":         key,
		"timeoutInMs": timeoutInMs,
	})
}

func testAccAPIExtensionConfig(identifier, key string, timeoutInMs int) string {
	return hclTemplate(`
		resource "commercetools_api_extension" "{{ .identifier }}" {
			key = "{{ .key }}"
			timeout_in_ms = {{ .timeoutInMs }}

			destination {
				type                 = "HTTP"
				url                  = "https://example.com"
				authorization_header = "Basic 12345"
			}

			trigger {
				resource_type_id = "customer"
				actions = ["Create"]
			}
		}
	`, map[string]any{
		"identifier":  identifier,
		"key":         key,
		"timeoutInMs": timeoutInMs,
	})
}

func testAccAPIExtensionConfigRequiredOnly(identifier, key string) string {
	return hclTemplate(`
		resource "commercetools_api_extension" "{{ .identifier }}" {
			key = "{{ .key }}"

			destination {
				type = "HTTP"
				url  = "https://example.com"
			}

			trigger {
				resource_type_id = "customer"
				actions = ["Create"]
			}
		}
	`, map[string]any{
		"identifier": identifier,
		"key":        key,
	})
}

func testAccAPIExtensionUpdate(identifier, key string, timeoutInMs int) string {
	return hclTemplate(`
		resource "commercetools_api_extension" "{{ .identifier }}" {
			key = "{{ .key }}"
			timeout_in_ms = {{ .timeoutInMs }}

			destination {
				type                 = "HTTP"
				url                  = "https://example.com"
				authorization_header = "Basic 12345"
			}

			trigger {
				resource_type_id = "customer"
				actions = ["Create", "Update"]
				condition = "name = \"Michael\""
			}
		}
	`, map[string]any{
		"identifier":  identifier,
		"key":         key,
		"timeoutInMs": timeoutInMs,
	})
}

func testAccAPIExtensionExists(n string) resource.TestCheckFunc {
	identifier := fmt.Sprintf("commercetools_api_extension.%s", n)
	return func(s *terraform.State) error {
		_, err := testGetExtension(s, identifier)
		return err
	}
}

func testAccCheckAPIExtensionDestroy(s *terraform.State) error {
	client := getClient(testAccProvider.Meta())

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "commercetools_api_extension" {
			continue
		}
		response, err := client.Extensions().WithId(rs.Primary.ID).Get().Execute(context.Background())
		if err == nil {
			if response != nil && response.ID == rs.Primary.ID {
				return fmt.Errorf("api extension (%s) still exists", rs.Primary.ID)
			}
			return nil
		}
		if newErr := checkApiResult(err); newErr != nil {
			return newErr
		}
	}
	return nil
}

func testGetExtension(s *terraform.State, identifier string) (*platform.Extension, error) {
	rs, ok := s.RootModule().Resources[identifier]
	if !ok {
		return nil, fmt.Errorf("API Extension %s not found", identifier)
	}

	client := getClient(testAccProvider.Meta())
	result, err := client.Extensions().WithId(rs.Primary.ID).Get().Execute(context.Background())
	if err != nil {
		return nil, err
	}
	return result, nil
}

func TestAccAPIExtension_azure_authentication(t *testing.T) {
	name := fmt.Sprintf("extension_%s", acctest.RandString(5))
	timeoutInMs := acctest.RandIntRange(200, 1800)
	identifier := "ext"
	resourceName := "commercetools_api_extension.ext"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAPIExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIExtensionAzureFunctionsConfig(identifier, name, timeoutInMs),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists("ext"),
					resource.TestCheckResourceAttr(
						resourceName, "key", name),
					resource.TestCheckResourceAttr(
						resourceName, "timeout_in_ms", strconv.FormatInt(int64(timeoutInMs), 10)),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "trigger.0.actions.0", "Create"),
				),
			},
			{
				Config:   testAccAPIExtensionAzureFunctionsConfig(identifier, name, timeoutInMs),
				PlanOnly: true,
			},
		},
	})
}

func testAccAPIExtensionAzureFunctionsConfig(identifier, key string, timeoutInMs int) string {
	return hclTemplate(`
		resource "commercetools_api_extension" "{{ .identifier }}" {
		  key = "{{ .key }}"
	      timeout_in_ms = {{ .timeoutInMs }}
		
		  destination {
			url                  = "http://google.com"
			azure_authentication = "my-other-auth-string"
			type                 = "HTTP"
		  }
		
		  trigger {
			resource_type_id = "customer"
			actions          = ["Create"]
		  }
		}
	`, map[string]any{
		"identifier":  identifier,
		"key":         key,
		"timeoutInMs": timeoutInMs,
	})
}

func TestAccAPIExtension_additionalContext(t *testing.T) {
	// The additional_context attribute is only exercised through update actions
	// here: the pinned commercetools mock server silently drops
	// additionalContext when set on create (fixed upstream, not yet released -
	// see https://github.com/labd/commercetools-node-mock/pull/430), but
	// setAdditionalContext was already implemented, so updates work correctly.
	name := fmt.Sprintf("extension_%s", acctest.RandString(5))
	identifier := "ext"
	resourceName := "commercetools_api_extension.ext"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckAPIExtensionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIExtensionConfigRequiredOnly(identifier, name),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "additional_context.#", "0"),
				),
			},
			{
				Config: testAccAPIExtensionAdditionalContextConfig(identifier, name, true),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "additional_context.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "additional_context.0.include_old_resource", "true"),
				),
			},
			{
				Config:   testAccAPIExtensionAdditionalContextConfig(identifier, name, true),
				PlanOnly: true,
			},
			{
				Config: testAccAPIExtensionAdditionalContextConfig(identifier, name, false),
				Check: resource.ComposeTestCheckFunc(
					testAccAPIExtensionExists(identifier),
					resource.TestCheckResourceAttr(
						resourceName, "additional_context.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "additional_context.0.include_old_resource", "false"),
				),
			},
		},
	})
}

func testAccAPIExtensionAdditionalContextConfig(identifier, key string, includeOldResource bool) string {
	return hclTemplate(`
		resource "commercetools_api_extension" "{{ .identifier }}" {
			key = "{{ .key }}"

			destination {
				type = "HTTP"
				url  = "https://example.com"
			}

			trigger {
				resource_type_id = "customer"
				actions = ["Create", "Update"]
			}

			additional_context {
				include_old_resource = {{ .includeOldResource }}
			}
		}
	`, map[string]any{
		"identifier":         identifier,
		"key":                key,
		"includeOldResource": includeOldResource,
	})
}
