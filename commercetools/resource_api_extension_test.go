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

func TestExpandExtensionDependencies(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{})
	assert.Nil(t, expandExtensionDependencies(d))

	d = schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{
		"dependencies": []any{
			"my-extension-key",
			"2845b936-e407-4f29-957b-f8deb0fcba97",
		},
	})

	dependencies := expandExtensionDependencies(d)
	assert.Len(t, dependencies, 2)

	// Anything that is not a UUID is sent as a key
	assert.Nil(t, dependencies[0].ID)
	assert.Equal(t, "my-extension-key", *dependencies[0].Key)

	// A UUID is sent as an id
	assert.Nil(t, dependencies[1].Key)
	assert.Equal(t, "2845b936-e407-4f29-957b-f8deb0fcba97", *dependencies[1].ID)
}

func TestExtensionDependencyIDs(t *testing.T) {
	assert.Equal(t, []string{}, extensionDependencyIDs(nil))
	assert.Equal(t,
		[]string{"2845b936-e407-4f29-957b-f8deb0fcba97", "7ba7f2b4-1f5d-4b9c-9d4a-1c0e6e0f9d21"},
		extensionDependencyIDs([]platform.ExtensionReference{
			{ID: "2845b936-e407-4f29-957b-f8deb0fcba97"},
			{ID: "7ba7f2b4-1f5d-4b9c-9d4a-1c0e6e0f9d21"},
		}))
}

func TestIsExtensionID(t *testing.T) {
	assert.True(t, isExtensionID("2845b936-e407-4f29-957b-f8deb0fcba97"))
	assert.False(t, isExtensionID("my-extension-key"))
	assert.False(t, isExtensionID(""))
	assert.False(t, isExtensionID("2845b936-e407-4f29-957b-f8deb0fcba9"))
}

func TestFlattenExtensionDependencies(t *testing.T) {
	const (
		firstID   = "2845b936-e407-4f29-957b-f8deb0fcba97"
		secondID  = "7ba7f2b4-1f5d-4b9c-9d4a-1c0e6e0f9d21"
		keylessID = "9c1f1d18-1b1f-4a1a-bd60-0b3f4b3f9b02"
	)
	keys := map[string]string{firstID: "first-extension", secondID: "second-extension"}

	withState := func(dependencies ...string) *schema.ResourceData {
		raw := make([]any, 0, len(dependencies))
		for _, dependency := range dependencies {
			raw = append(raw, dependency)
		}
		return schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{
			"dependencies": raw,
		})
	}

	// Configured by key: the key is kept, so no diff is introduced
	assert.Equal(t,
		[]string{"first-extension", "second-extension"},
		flattenExtensionDependencies(
			[]string{firstID, secondID}, keys, withState("first-extension", "second-extension")))

	// Configured by id: the id is kept
	assert.Equal(t,
		[]string{firstID, secondID},
		flattenExtensionDependencies([]string{firstID, secondID}, keys, withState(firstID, secondID)))

	// Both notations can be mixed
	assert.Equal(t,
		[]string{firstID, "second-extension"},
		flattenExtensionDependencies(
			[]string{firstID, secondID}, keys, withState(firstID, "second-extension")))

	// Not in the state yet (import or drift): prefer the key
	assert.Equal(t,
		[]string{"first-extension"},
		flattenExtensionDependencies([]string{firstID}, keys, withState()))

	// An extension without a key can only be written as an id
	assert.Equal(t,
		[]string{keylessID},
		flattenExtensionDependencies([]string{keylessID}, keys, withState()))

	assert.Equal(t, []string{}, flattenExtensionDependencies(nil, keys, withState()))
}

func TestExpandExtensionExpansionPaths(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{})
	assert.Nil(t, expandExtensionExpansionPaths(d))

	d = schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{
		"expansion_paths": []any{"lineItems[*].variant", "customerGroup"},
	})
	assert.Equal(t,
		[]string{"lineItems[*].variant", "customerGroup"},
		expandExtensionExpansionPaths(d))
}

func TestExpandExtensionAdditionalContext(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{})
	assert.Nil(t, expandExtensionAdditionalContext(d))

	d = schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{
		"additional_context": []any{
			map[string]any{"include_old_resource": true},
		},
	})

	additionalContext := expandExtensionAdditionalContext(d)
	assert.NotNil(t, additionalContext)
	assert.Equal(t, true, *additionalContext.IncludeOldResource)
}

func TestFlattenExtensionAdditionalContext(t *testing.T) {
	empty := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{})
	configured := schema.TestResourceDataRaw(t, resourceAPIExtension().Schema, map[string]any{
		"additional_context": []any{
			map[string]any{"include_old_resource": false},
		},
	})

	// Not returned by commercetools at all
	assert.Empty(t, flattenExtensionAdditionalContext(nil, empty))

	// Returned with the default value while never configured: keep it absent so
	// we don't introduce a permanent diff.
	assert.Empty(t, flattenExtensionAdditionalContext(
		&platform.ExtensionAdditionalContext{IncludeOldResource: false}, empty))

	// Returned with the default value while explicitly configured: keep it.
	assert.Equal(t,
		[]map[string]any{{"include_old_resource": false}},
		flattenExtensionAdditionalContext(
			&platform.ExtensionAdditionalContext{IncludeOldResource: false}, configured))

	// A non-default value is always written
	assert.Equal(t,
		[]map[string]any{{"include_old_resource": true}},
		flattenExtensionAdditionalContext(
			&platform.ExtensionAdditionalContext{IncludeOldResource: true}, empty))
}
