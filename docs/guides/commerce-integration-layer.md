---
subcategory: ""
page_title: "Commerce Integration Layer"
description: |-
    A full example of setting up a Project on the commercetools Commerce Integration Layer
---

# Commerce Integration Layer

The Commerce Integration Layer (CIL) serves a GraphQL API for storefronts on
top of a Project. Its core covers the catalog, carts, customers and checkout;
an extension bundle adds a project's own fields and the services it calls,
such as a CMS or a search engine.

A Project on CIL has three parts, set up in different ways:

| Part | How |
| --- | --- |
| Enrolling the Project in CIL | Once, in the Merchant Center. There is no API for it. |
| The Project's settings: session defaults, the hosts the bundle may call, its config and secrets, hidden fields, login and checkout | Terraform, with the `commercetools_integration_layer_*` resources in this guide. |
| The extension bundle | Built and pushed from CI/CD with the commercetools CLI: `commercetools integration-layer extension push dist/extension.cjs`. Each push is a new version; Terraform does not manage it. |

This guide sets up the settings for a bundle that reads content from
Contentful, searches with Algolia, suggests addresses with Loqate and reads
Custom Objects from the Project, and checks out with commercetools Checkout.

## The provider

The CIL management API is on a host of its own. The provider derives it from
the region in `api_url`, such as
`https://extensions.integration-layer.europe-west1.gcp.commercetools.com`; set
`integration_layer_url` only to use another host. The API client needs the
`manage_project` scope.

```hcl
terraform {
  required_providers {
    commercetools = {
      source = "labd/commercetools"
    }
  }
}

provider "commercetools" {
  client_id     = var.ctp_client_id
  client_secret = var.ctp_client_secret
  project_key   = var.project_key
  scopes        = "manage_project:${var.project_key}"
  api_url       = "https://api.europe-west1.gcp.commercetools.com"
  token_url     = "https://auth.europe-west1.gcp.commercetools.com"
}

variable "project_key" {
  type = string
}

variable "ctp_client_id" {
  type = string
}

variable "ctp_client_secret" {
  type      = string
  sensitive = true
}
```

## Session defaults

A shopper's session starts with the Project's default language, currency and
country. Each has to be one the Project offers.

```hcl
resource "commercetools_integration_layer_session_defaults" "defaults" {
  label    = "My Shop"
  language = "nl-NL"
  currency = "EUR"
  country  = "NL"
}
```

## The hosts the bundle may call

The bundle can only call hosts on the allowlist. List each host its modules
call; commercetools denies some hosts whatever the list says, and the
resource's `deny` shows which.

```hcl
resource "commercetools_integration_layer_allowlist" "allowlist" {
  allow = [
    "cdn.contentful.com",
    "preview.contentful.com",
    "*.algolia.net",
    "*.algolianet.com",
    "api.addressy.com",
    # The platform API, for the module that reads Custom Objects.
    "auth.europe-west1.gcp.commercetools.com",
    "api.europe-west1.gcp.commercetools.com",
  ]
}
```

## Config and secrets

The bundle reads its configuration as `ctx.config.<KEY>`. A plain `value` can
be read back by anyone who can read the Project's config; a `secret_value` is
stored as a secret and never returned. Each key is a resource of its own, so
keys set outside of Terraform, such as with the CLI, are left alone.

The module that reads Custom Objects gets an API client of its own, with only
the scope it needs, whose credentials go into the config.

```hcl
resource "commercetools_api_client" "extension" {
  name  = "cil-extension"
  scope = ["view_key_value_documents:${var.project_key}"]
}

variable "secrets" {
  type = object({
    contentful_access_token  = string
    contentful_preview_token = string
    algolia_search_api_key   = string
    loqate_api_key           = string
  })
  sensitive = true
}

locals {
  config = {
    CONTENTFUL_SPACE_ID    = "my-space"
    CONTENTFUL_ENVIRONMENT = "master"
    ALGOLIA_APP_ID         = "MYAPPID"
    LOQATE_COUNTRIES       = "NL,BE,DE"
    CTP_PROJECT_KEY        = var.project_key
    CTP_CLIENT_ID          = commercetools_api_client.extension.id
  }

  # Keyed by name alone, so the for_each can be planned without the secrets.
  secret_keys = toset([
    "CONTENTFUL_ACCESS_TOKEN",
    "CONTENTFUL_PREVIEW_TOKEN",
    "ALGOLIA_SEARCH_API_KEY",
    "LOQATE_API_KEY",
    "CTP_CLIENT_SECRET",
  ])
}

resource "commercetools_integration_layer_config_entry" "config" {
  for_each = local.config
  key      = each.key
  value    = each.value
}

resource "commercetools_integration_layer_config_entry" "secret" {
  for_each = local.secret_keys
  key      = each.key
  secret_value = {
    CONTENTFUL_ACCESS_TOKEN  = var.secrets.contentful_access_token
    CONTENTFUL_PREVIEW_TOKEN = var.secrets.contentful_preview_token
    ALGOLIA_SEARCH_API_KEY   = var.secrets.algolia_search_api_key
    LOQATE_API_KEY           = var.secrets.loqate_api_key
    CTP_CLIENT_SECRET        = commercetools_api_client.extension.secret
  }[each.key]
}
```

## Hidden fields

Custom fields and product attributes are in the GraphQL schema unless they are
hidden. Hide the ones a storefront must not see.

```hcl
resource "commercetools_integration_layer_field_visibility" "visibility" {
  hidden_field {
    space      = "attribute"
    type_key   = "chair"
    field_name = "supplierCode"
  }

  hidden_field {
    space      = "customField"
    type_key   = "order-extra"
    field_name = "costPrice"
  }
}
```

## Login with an identity provider

Shoppers can sign in with an OpenID Connect identity provider. A signed-in
shopper is matched to a Customer by email address or by external id. The
client secret is only sent; after the first apply it can be left out to keep
the stored one.

```hcl
variable "idp_client_secret" {
  type      = string
  sensitive = true
}

resource "commercetools_integration_layer_idp_login" "login" {
  issuer              = "https://login.example.com"
  token_endpoint      = "https://login.example.com/oauth2/token"
  revocation_endpoint = "https://login.example.com/oauth2/revoke"
  jwks_uri            = "https://login.example.com/.well-known/jwks.json"
  client_id           = "storefront"
  client_secret       = var.idp_client_secret
  redirect_uri        = "https://shop.example.com/auth/callback"
  match_strategy      = "email"
}
```

## Checkout

CIL checks out with commercetools Checkout. The checkout rules pick the
Checkout application for a cart, by its country or Store, in order; a rule
without either matches every cart and comes last. The applications and their
payment methods are set up as in the [Checkout with Connect](checkout) guide.

```hcl
resource "commercetools_checkout_application" "shop" {
  key                           = "shop"
  name                          = "Shop"
  mode                          = "PaymentOnly"
  status                        = "Active"
  countries                     = ["NL", "BE", "DE"]
  allowed_origins               = ["https://shop.example.com"]
  active_payment_component_type = "Component"
}

resource "commercetools_integration_layer_checkout_rules" "rules" {
  rule {
    application_key = commercetools_checkout_application.shop.key
    mode            = "PAYMENT_ONLY"
  }
}
```

## Pushing the bundle

With the settings in place, CI/CD builds the bundle and pushes it:

```sh
commercetools integration-layer extension push dist/extension.cjs
```

Push after `terraform apply`, so the hosts and config keys a new version of the
bundle needs are there when it starts. When a version stops needing a key or a
host, remove it from Terraform after the push.

## Importing

Every resource in this guide except `commercetools_integration_layer_config_entry`
holds a setting the Project has once, and is imported by the Project key:

```sh
terraform import commercetools_integration_layer_session_defaults.defaults my-project
```

A config entry is imported by its key:

```sh
terraform import 'commercetools_integration_layer_config_entry.config["ALGOLIA_APP_ID"]' ALGOLIA_APP_ID
```

A secret's value is never returned, so an imported secret keeps its stored
value until Terraform writes `secret_value`. The same goes for the IdP login's
client secret.
