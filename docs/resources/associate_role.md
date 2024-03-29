---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "commercetools_associate_role Resource - terraform-provider-commercetools"
subcategory: ""
description: |-
  Associate Roles provide a way to group granular Permissions and assign them to Associates within a Business Unit.
  See also the Associate Role API Documentation https://docs.commercetools.com/api/projects/associate-roles
---

# commercetools_associate_role (Resource)

Associate Roles provide a way to group granular Permissions and assign them to Associates within a Business Unit.

See also the [Associate Role API Documentation](https://docs.commercetools.com/api/projects/associate-roles)

## Example Usage

```terraform
resource "commercetools_associate_role" "regional_manager" {
  key              = "regional-manager-europe"
  buyer_assignable = false
  name             = "Regional Manager - Europe"
  permissions = [
    "AddChildUnits",
    "UpdateBusinessUnitDetails",
    "UpdateAssociates",
    "CreateMyCarts",
    "DeleteMyCarts",
    "UpdateMyCarts",
    "ViewMyCarts",
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `key` (String) User-defined unique identifier of the AssociateRole.
- `permissions` (List of String) List of Permissions for the AssociateRole.

### Optional

- `buyer_assignable` (Boolean) Whether the AssociateRole can be assigned to an Associate by a buyer. If false, the AssociateRole can only be assigned using the general endpoint.
- `name` (String) Name of the AssociateRole.

### Read-Only

- `id` (String) Unique identifier of the AssociateRole.
- `version` (Number) Current version of the AssociateRole.
