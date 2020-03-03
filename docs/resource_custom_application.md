# Custom Application (Merchant Center)

Manage the resources for [Custom Applications](https://docs.commercetools.com/custom-applications).

## Example Usage

Assuming you are building a Custom Application to manage State Machines, you can provide the following configuration:

```hcl
resource "commercetools_custom_application" "state-machines" {
  name        = "State machines"
  description = "Manage state machines"
  url         = "https://state-machines.my-domain.com"
  is_active   = true
  navbar_menu {
    uri_path = "state-machines"
    icon     = "RocketIcon"
    label_all_locales {
      locale = "en"
      value  = "State machines"
    }
    label_all_locales {
      locale = "de"
      value  = "Zustandsmachinen"
    }
    permissions = ["ViewDeveloperSettings"]

    submenu {
      uri_path = "state-machines/new"
      label_all_locales {
        locale = "en"
        value  = "Add state machine"
      }
      label_all_locales {
        locale = "de"
        value  = "Zustandsmachine hinzufügen"
      }
      permissions = ["ManageDeveloperSettings"]
    }
    # submenu {
    #   # ...
    # }
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - The name of the Custom Application.
* `description` (_optional) - The description of the Custom Application.
* `url` - The URL (origin) of the Custom Application. The Merchant Center serves Custom Applications on its own domain, but requests are internally forwarded to this URL.
* `is_active` - Whether to activate or deactivate the Custom Application.
* `navbar_menu` - The [Navbar Menu](#navbar-menu) configuration for the links in the main navigation.

### Navbar Menu

* `uri_path` - The main route path forms the URL for the Custom Application in the Merchant Center. The URL always contains the project key followed by the value in this field. The main route path matches requests forwarded to the application. Do not use spaces in your URL, use dash separators.
* `icon` - The icon to be shown in the navigation menu on the left side.
* `permissions` - A list of permission strings to be applied to the navigation links. If a user does not have at least one of the permissions, they won't be able to see the Custom Application in the Merchant Center navigation menu.
* `label_all_locales` - A list of localized labels. The values are used in the navigation links according to the language/locale settings in the user profile.
* `submenu` - The [Submenu](#submenu) configuration for the links in the main navigation.

### Submenu

* `uri_path` - The main route path forms the URL for the Custom Application in the Merchant Center. The URL always contains the project key followed by the value in this field. The main route path matches requests forwarded to the application. Do not use spaces in your URL, use dash separators.
* `permissions` - A list of permission strings to be applied to the navigation links. If a user does not have at least one of the permissions, they won't be able to see the Custom Application in the Merchant Center navigation menu.
* `label_all_locales` - A list of localized labels. The values are used in the navigation links according to the language/locale settings in the user profile.
