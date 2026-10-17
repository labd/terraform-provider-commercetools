---
subcategory: ""
page_title: "State Import"
description: |-
    Bringing existing commercetools resources under Terraform management
---

# State Import

Resources that already exist in a Project, created in the Merchant Center or
by a script, can be brought under Terraform management by importing them. After
an import, Terraform manages the resource as if it had created it.

What a resource is imported by differs per resource: most by their id, some
by their key, and the Project-wide settings of the Commerce Integration Layer
by the Project key. The Import section of each resource's page says which.

## With an import block

Since Terraform 1.5, an `import` block declares the import in the
configuration, so it is planned and reviewed like any other change:

```hcl
import {
  to = commercetools_category.chairs
  id = "21abd75b-1506-412f-a404-4e1228974c3a"
}

resource "commercetools_category" "chairs" {
  key = "chairs"
  name = {
    en = "Chairs"
  }
  slug = {
    en = "chairs"
  }
}
```

`terraform plan` then shows the import and every difference between the
configuration and the category as it is. Change the configuration until the
plan only imports, then apply. The `import` block can be removed afterwards.

Terraform can also write the configuration for you:

```sh
terraform plan -generate-config-out=generated.tf
```

With an `import` block and no `resource` block for it, this writes the
resource as it is to `generated.tf`, to review and move into your
configuration.

## With the import command

The `terraform import` command imports one resource into the state right away,
without a plan:

```sh
terraform import commercetools_category.chairs 21abd75b-1506-412f-a404-4e1228974c3a
```

The resource block has to be in the configuration first. Run `terraform plan`
after the import to see where the configuration and the resource differ:

```
  # commercetools_category.chairs will be updated in-place
  ~ resource "commercetools_category" "chairs" {
        id         = "21abd75b-1506-412f-a404-4e1228974c3a"
      ~ name       = {
          ~ "en" = "Seating" -> "Chairs"
        }
      - order_hint = "0.1" -> null
        # (2 unchanged attributes hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

Here the category's name differs and its order hint is missing from the
configuration. Either change the configuration to match, or apply to change
the category to match the configuration.

## Secrets

The API never returns secrets, such as the secured configuration of a Connect
deployment, a secret config entry of the Commerce Integration Layer or the
client secret of its IdP login. After an import they are unknown to Terraform:
set them in the configuration, and the next apply writes them.
