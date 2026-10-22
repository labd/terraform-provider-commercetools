---
subcategory: ""
page_title: "Checkout with Connect"
description: |-
    Setting up commercetools Checkout with the Adyen and Voucherify connectors on commercetools Connect
---

# Checkout with Connect

[commercetools Checkout](https://docs.commercetools.com/checkout) runs the
checkout, or only its payment step, for a storefront. It does not talk to a
payment service provider (PSP) itself: each payment method is handled by a
connector that runs on [commercetools Connect](https://docs.commercetools.com/connect).

This guide sets up Checkout with two public connectors from the Connect
marketplace, both by commercetools:

- [Adyen for Checkout](https://github.com/commercetools/connect-payment-integration-adyen),
  for cards, iDEAL and PayPal;
- [Voucherify Giftcard for Checkout](https://github.com/commercetools/connect-giftcard-integration-voucherify),
  for paying with a gift card.

It takes these resources, each referring to the ones before:

| Resource | What it is |
| --- | --- |
| `commercetools_api_client` | An API client for each connector, which it calls the Project with. |
| `commercetools_connect_deployment` | A connector installed in the Project, with its configuration and the PSP's credentials. |
| `commercetools_checkout_application` | How Checkout runs for a storefront: countries, origins and the payment component. |
| `commercetools_checkout_payment_integration` | A payment method of the application, handled by a deployment. |

A connector of your own works the same way, with a
`commercetools_connect_connector` for it; see [A connector of your own](#a-connector-of-your-own).

## Before you start

- Checkout has to be activated for the Project in the Merchant Center.
- Using Adyen needs a contract with Adyen, and the merchant account, API key and
  client key from its Customer Area. Using Voucherify needs a Voucherify
  project, with its application ID and secret key.
- The API client Terraform runs with needs these scopes, next to the ones for
  the rest of your configuration:
  - `manage_api_clients:{projectKey}`, for the connectors' API clients,
  - `manage_connectors_deployments:{projectKey}`, for the deployments,
  - `manage_checkout_applications:{projectKey}` and
    `manage_checkout_payment_integrations:{projectKey}`, for Checkout.
- Connect and Checkout have hosts of their own. The provider derives them from
  the region in `api_url`, such as `https://connect.europe-west1.gcp.commercetools.com`
  and `https://checkout.europe-west1.gcp.commercetools.com`. Set `connect_url`
  and `checkout_url` only to use other hosts.

```hcl
terraform {
  required_providers {
    commercetools = {
      source = "labd/commercetools"
    }
  }
}

locals {
  project_key = "my-project"
  region      = "europe-west1.gcp"
}

provider "commercetools" {
  client_id     = var.ctp_client_id
  client_secret = var.ctp_client_secret
  project_key   = local.project_key
  scopes = join(" ", [for scope in [
    "manage_project",
    "manage_api_clients",
    "manage_connectors_deployments",
    "manage_checkout_applications",
    "manage_checkout_payment_integrations",
  ] : "${scope}:${local.project_key}"])
  api_url   = "https://api.${local.region}.commercetools.com"
  token_url = "https://auth.${local.region}.commercetools.com"
}

variable "ctp_client_id" {
  type = string
}

variable "ctp_client_secret" {
  type      = string
  sensitive = true
}
```

## 1. The connectors' API clients

Each connector calls the Project with an API client of its own, with the
scopes its `connect.yaml` lists.

```hcl
resource "commercetools_api_client" "adyen" {
  name = "checkout-adyen"
  scope = [for scope in [
    "manage_payments",
    "manage_orders",
    "view_sessions",
    "view_api_clients",
    "introspect_oauth_tokens",
  ] : "${scope}:${local.project_key}"]
}

resource "commercetools_api_client" "voucherify" {
  name = "checkout-voucherify"
  scope = [for scope in [
    "manage_payments",
    "manage_orders",
    "view_sessions",
    "view_api_clients",
    "manage_checkout_payment_intents",
    "introspect_oauth_tokens",
  ] : "${scope}:${local.project_key}"]
}
```

## 2. The deployments

A public connector is deployed by the key the Connect marketplace lists for it,
without a `commercetools_connect_connector`. Leave `connector_version` out to
deploy its current version.

Both connectors have two applications: `enabler`, the frontend component
Checkout loads, and `processor`, the service that talks to the PSP. Only the
processor has configuration. Plain values go in `standard_configuration`,
secrets in `secured_configuration`; each connector's `connect.yaml` lists its
keys, with the optional ones left out here.

Creating a deployment waits until Connect has built and deployed it, which
can take up to 15 minutes.

```hcl
variable "adyen_connector_key" {
  description = "Key of Adyen for Checkout, as the Connect marketplace lists it"
  type        = string
}

variable "voucherify_connector_key" {
  description = "Key of Voucherify Giftcard for Checkout, as the Connect marketplace lists it"
  type        = string
}

variable "adyen" {
  type = object({
    merchant_account  = string
    client_key        = string
    api_key           = string
    notification_hmac = string
  })
  sensitive = true
}

variable "voucherify" {
  type = object({
    application_id = string
    secret_key     = string
  })
  sensitive = true
}

locals {
  # The commercetools hosts both connectors are configured with.
  ctp_configuration = {
    CTP_PROJECT_KEY = local.project_key
    CTP_AUTH_URL    = "https://auth.${local.region}.commercetools.com"
    CTP_API_URL     = "https://api.${local.region}.commercetools.com"
    CTP_SESSION_URL = "https://session.${local.region}.commercetools.com"
    CTP_JWKS_URL    = "https://mc-api.${local.region}.commercetools.com/.well-known/jwks.json"
    CTP_JWT_ISSUER  = "https://mc-api.${local.region}.commercetools.com"
  }
}

resource "commercetools_connect_deployment" "adyen" {
  key           = "checkout-adyen"
  connector_key = var.adyen_connector_key
  region        = local.region
  type          = "production"

  application {
    name = "enabler"
  }

  application {
    name = "processor"
    standard_configuration = merge(local.ctp_configuration, {
      CTP_CLIENT_ID          = commercetools_api_client.adyen.id
      CTP_CHECKOUT_URL       = "https://checkout.${local.region}.commercetools.com"
      ADYEN_ENVIRONMENT      = "TEST"
      ADYEN_MERCHANT_ACCOUNT = var.adyen.merchant_account
      ADYEN_CLIENT_KEY       = var.adyen.client_key
      MERCHANT_RETURN_URL    = "https://shop.example.com/checkout/return"
    })
    secured_configuration = {
      CTP_CLIENT_SECRET           = commercetools_api_client.adyen.secret
      ADYEN_API_KEY               = var.adyen.api_key
      ADYEN_NOTIFICATION_HMAC_KEY = var.adyen.notification_hmac
    }
  }
}

resource "commercetools_connect_deployment" "voucherify" {
  key           = "checkout-voucherify"
  connector_key = var.voucherify_connector_key
  region        = local.region
  type          = "production"

  application {
    name = "enabler"
  }

  application {
    name = "processor"
    standard_configuration = merge(local.ctp_configuration, {
      CTP_CLIENT_ID             = commercetools_api_client.voucherify.id
      VOUCHERIFY_APPLICATION_ID = var.voucherify.application_id
      VOUCHERIFY_API_URL        = "https://api.voucherify.io"
      VOUCHERIFY_CURRENCY       = "EUR"
    })
    secured_configuration = {
      CTP_CLIENT_SECRET     = commercetools_api_client.voucherify.secret
      VOUCHERIFY_SECRET_KEY = var.voucherify.secret_key
    }
  }
}
```

`ADYEN_ENVIRONMENT` is `TEST` or `LIVE`. A live account also needs
`ADYEN_LIVE_URL_PREFIX`.

### Adyen's webhook

Adyen reports the outcome of a payment to the processor with a webhook. Set it
up in Adyen's Customer Area with the processor's URL followed by
`/notifications`:

```hcl
output "adyen_webhook_url" {
  value = "${one([
    for app in commercetools_connect_deployment.adyen.application : app.url if app.name == "processor"
  ])}/notifications"
}
```

The HMAC key that signs the webhook is only known once the webhook exists in
Adyen. Deploy first with a placeholder for `notification_hmac`, create the
webhook, then set the real key and apply again: a changed secret redeploys the
deployment in place.

## 3. The Checkout application

The application is what a storefront loads Checkout with, by its key. It lists
the countries it serves and the origins allowed to load it. `CompleteFlow` runs
the whole checkout and needs a logo; `PaymentOnly` runs only the payment step,
after the storefront's own checkout.

```hcl
resource "commercetools_checkout_application" "shop" {
  key                           = "shop"
  name                          = "Shop"
  mode                          = "CompleteFlow"
  status                        = "Active"
  countries                     = ["NL", "BE"]
  allowed_origins               = ["https://shop.example.com"]
  logo_url                      = "https://shop.example.com/logo.svg"
  payment_return_url            = "https://shop.example.com/checkout/return"
  active_payment_component_type = "Component"
  allow_discounts               = true

  agreement {
    name = "terms"
    type = "MandatoryCheckbox"
    text = {
      en = "I agree to the terms and conditions."
      nl = "Ik ga akkoord met de algemene voorwaarden."
    }
  }
}
```

## 4. The payment integrations

Each payment method is a payment integration of the application, handled by a
deployment. Its `type` is the payment method as the connector names it: for
Adyen's components `card`, `ideal`, `paypal`, `applepay`, `googlepay`,
`klarna_pay_later` and more, for Voucherify `giftcard`. Checkout offers only
the integrations whose `component_type` matches the application's
`active_payment_component_type`, in the order of `sort_priority`, and of
those only the ones whose `predicate` matches the cart.

```hcl
resource "commercetools_checkout_payment_integration" "card" {
  key                   = "adyen-card"
  application_id        = commercetools_checkout_application.shop.id
  connect_deployment_id = commercetools_connect_deployment.adyen.id
  name                  = "Adyen card"
  type                  = "card"
  component_type        = "Component"
  status                = "Active"
  sort_priority         = 1

  label = {
    en = "Credit card"
    nl = "Creditcard"
  }

  # Reverse the payment when the order cannot be placed.
  automated_reversal = "Active"
}

resource "commercetools_checkout_payment_integration" "ideal" {
  key                   = "adyen-ideal"
  application_id        = commercetools_checkout_application.shop.id
  connect_deployment_id = commercetools_connect_deployment.adyen.id
  name                  = "Adyen iDEAL"
  type                  = "ideal"
  component_type        = "Component"
  status                = "Active"
  predicate             = "country='NL'"
  sort_priority         = 2

  label = {
    en = "iDEAL"
  }
}

resource "commercetools_checkout_payment_integration" "paypal" {
  key                   = "adyen-paypal"
  application_id        = commercetools_checkout_application.shop.id
  connect_deployment_id = commercetools_connect_deployment.adyen.id
  name                  = "Adyen PayPal"
  type                  = "paypal"
  component_type        = "Component"
  status                = "Active"
  sort_priority         = 3

  label = {
    en = "PayPal"
  }
}

# A gift card pays part of the order at most, so it is offered next to the
# other payment methods, for the rest.
resource "commercetools_checkout_payment_integration" "giftcard" {
  key                   = "voucherify-giftcard"
  application_id        = commercetools_checkout_application.shop.id
  connect_deployment_id = commercetools_connect_deployment.voucherify.id
  name                  = "Voucherify gift card"
  type                  = "giftcard"
  component_type        = "Component"
  status                = "Active"
  sort_priority         = 10

  label = {
    en = "Gift card"
    nl = "Cadeaukaart"
  }
}
```

## Using it from the Commerce Integration Layer

With the [Commerce Integration Layer](commerce-integration-layer), the checkout
rules decide which application a cart checks out with:

```hcl
resource "commercetools_integration_layer_checkout_rules" "rules" {
  rule {
    application_key = commercetools_checkout_application.shop.key
    mode            = "COMPLETE"
  }
}
```

## Updating a public connector

A deployment keeps running the version of the connector it was deployed with.
Move it to the connector's latest version from CI/CD with the
[commercetools CLI](https://docs.commercetools.com/connect/cli):

```sh
commercetools connect deployment redeploy --key checkout-adyen --updateConnector
```

Terraform keeps the version that runs, as `connector_version` is left out, and
keeps managing the configuration: a changed value or secret redeploys the
deployment in place, with that version.

A redeploy by the CLI keeps the configuration values the deployment has. When
a new version needs a configuration key the running version does not know,
Connect rejects that key until the new version is deployed: pass it to the
CLI's redeploy, as `--configuration processor.NEW_KEY=value`, and then add it to
the deployment in Terraform.

## A connector of your own

A connector of your own, such as a fork of one of these, is a GitHub repository
with a `connect.yaml`, which Connect builds from a tag. For a private
repository, give its SSH URL and grant the `connect-mu` GitHub user read access
to it. The Terraform API client also needs `manage_connectors:{projectKey}`.

```hcl
resource "commercetools_connect_connector" "payments" {
  key               = "my-adyen"
  name              = "Adyen for Checkout, customized"
  repository_url    = "git@github.com:my-org/connect-payment-integration-adyen.git"
  repository_tag    = "v1.0.0"
  integration_types = ["psp"]
  supported_regions = [local.region]
  private_projects  = ["${local.region}:${local.project_key}"]

  # Publish it privately, so it can be deployed. A private connector needs no
  # certification.
  publish = true

  creator {
    email   = "dev@example.com"
    company = "Example Company"
  }

  lifecycle {
    # Releases set a new tag with the commercetools CLI.
    ignore_changes = [repository_tag]
  }
}
```

Its deployment refers to it, and records the version Terraform published:

```hcl
resource "commercetools_connect_deployment" "my_payments" {
  key               = "checkout-my-adyen"
  connector_key     = commercetools_connect_connector.payments.key
  connector_version = commercetools_connect_connector.payments.published_version
  region            = local.region
  type              = "production"

  application {
    name = "enabler"
  }

  application {
    name = "processor"
    # ...
  }

  lifecycle {
    # Releases are rolled out with the commercetools CLI.
    ignore_changes = [connector_version]
  }
}
```

A release is rolled out from CI/CD:

```sh
commercetools connect connectorstaged update --key my-adyen --repository-tag v1.1.0
commercetools connect connectorstaged publish --key my-adyen
# wait until the connector's status is Published, then:
commercetools connect deployment redeploy --key checkout-my-adyen --updateConnector
```

`ignore_changes` on the connector's `repository_tag` and the deployment's
`connector_version` keeps Terraform from moving them back.
