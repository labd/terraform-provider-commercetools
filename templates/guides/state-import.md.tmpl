---
subcategory: ""
page_title: "State Import"
---

This provider also allows for importing of existing resources into Terraform.
This can be useful to bring your existing commercetools resources under
Terraform management.

The [`import command`](https://developer.hashicorp.com/terraform/cli/import) is
used to import existing resources into Terraform. To achieve this the resource
needs to first be added to the terraform configuration. Subsequently,
the `import` command can be used to import the resource ID into state. When
running the next `terraform plan` or `terraform apply` the resource will be
updated to reflect the configuration by Terraform.

## Example

Assuming you have a commercetools category you want to import into Terraform:

```json
{
  "id": "21abd75b-1506-412f-a404-4e1228974c3a",
  "name": {
    "en": "My unexpected category name"
  },
  "slug": {
    "en": "my-category"
  },
  "orderHint": "1"
  // ... etcetera
}
```

The first step is to add the resource to the Terraform configuration:

```hcl
resource "commercetools_category" "my-category" {
  key = "my-category"
  name {
    en = "My category name"
  }
  slug {
    en = "my-category"
  }
}
```

Then you can run the import command:

`terraform import commercetools_category.my-category 21abd75b-1506-412f-a404-4e1228974c3a`

This will import the category id into Terraform state. However, at this point we
still want to check if the configuration matches the actual state of the
category.

To do this, run `terraform plan`, which in our case will show the following:

```bash
Terraform will perform the following actions:

  # commercetools_category.my-category will be updated in-place
  ~ resource "commercetools_category" "my-category" {
        id         = "21abd75b-1506-412f-a404-4e1228974c3a"
      ~ name       = {
          ~ "en" = "My unexpected category name" -> "My category name"
        }
      - order_hint = "1" -> null
      + slug       = {
          + "en" = "my-category"
        }
        # (1 unchanged attribute hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.

```

We can see that the name in the configuration is different from the actual
state, and the order hint is missing entirely. To fix this, we can either choose
to update the configuration, or allow Terraform to update Commercetools to match
the intended configuration.

In the above case we don't want to change the name, but we want to add the order 
hint to the configuration:

```hcl
resource "commercetools_category" "my-category" {
  name = {
    en = "My unexpected category name"
  }
  order_hint = "1"
  slug = {
    en = "my-category"
  }
}
```

Now `terraform` plan will show the following:

```bash
Terraform will perform the following actions:

  # commercetools_category.my-category will be updated in-place
  ~ resource "commercetools_category" "my-category" {
        id         = "21abd75b-1506-412f-a404-4e1228974c3a"
        name       = {
            "en" = "My unexpected category name"
        }
      + slug       = {
          + "en" = "my-category"
        }
        # (2 unchanged attributes hidden)
    }

Plan: 0 to add, 1 to change, 0 to destroy.
```

Once we are satisfied with the plan, we can run `terraform apply` to apply the
changes and store the intended state.

Now the category is under Terraform management.
