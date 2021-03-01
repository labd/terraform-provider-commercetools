v0.27.0 (2021-03-01)
====================
 - Resource Project: Add new `carts` field to documentation
 - Previous feature addition should have been minor update so making it minor now

v0.26.2 (2021-03-01)
====================
 - Resource Project: Add `carts` field with countryTaxRateFallBackEnabled setting 
 - Resource Project: Fix updating of `messages` field to explicitly set `false` when deleted or set to false in terraform
instead of relying on commercetools default settings for project in these scenarios 

v0.26.1 (2021-01-21)
====================
 - Api Extension Resource: Fixed typo in `trigger` field name that caused updates to actions in triggers to fail

v0.26.0 (2021-01-12)
====================
 - Add Customer Group resource (#141)
 - Type Resource: Allow updating the label of an existing Enum value
 - Type Resource: Add support to update a set of enum in a custom type
 - Fix ProductType and DiscountCode tests with real commercetools environment

v0.25.3 (2020-12-17)
====================
 - Force creation of new resource when changing a Store key. There is
   no update action for this available.

v0.25.2 (2020-12-17)
====================
 - Implement changing key of an existing channel resource

v0.25.1 (2020-12-05)
====================
 - Fix a bug in `resource_type` when the input_hint of a field was modified.

v0.25.0 (2020-11-27)
====================
 - Add `commercetools_custom_object` resource

v0.24.1 (2020-11-13)
====================
 - Add a workaround in tax_rate resources to handle an issue with changing
   id's after a tax category update

v0.24.0 (2020-11-13)
====================
 - Add support for managing supply channels in stores
 - Handle non-existing tax rates in tax categories when refreshing state

v0.23.0 (2020-08-28)
====================
 - New tag naming scheme (prefix with v) to conform to terraform repository
   requirements
 - Update terraform-plugin-sdk for compatibility with terraform 0.13

0.22.1 (2020-07-21)
===================
- Fix edge case where distribution channels were not updated

0.22.0 (2020-07-20)
===================
- Update commercetools-go-sdk dependency to v0.2.0. This version now properly
  handles oauth2 authentication failures (#117)
- resource `commercetools_store`: add distribution\_channels

0.21.2 (2020-06-11)
===================
- Add store languages attribute to `store` resource

0.21.1 (2020-04-22)
===================
- Fix channel read null pointer exception. Name should be optional.

0.21.0 (2020-02-27)
===================
- Provider arguments (`client_id`, `client_secret`, `project_key`,
  `scopes`, `token_url` and `api_url`) are now required
- Don't retry various calls if Commercetools returns an error (resulting in
  unnecessary retries/waiting times).
- resource `commercetools_api_client`: updating now recreates the resource since
  it cannot be updated.
- Dependency update: use terraform-plugin-sdk 1.7.0

0.20.0 (2020-02-22)
===================
- Add Azure Event Grid subscription

0.19.0 (2019-10-02)
===================
- Update all dependencies (use go 1.13, switch to terraform plugin sdk)

0.18.3 (2019-09-11)
===================
- Use Terraform 0.12.8 dependency

0.18.2 (2019-09-10)
===================
- Brew release Linux has incorrect artifact name
- Set GOPROXY for possible unreachable go packages

0.18.1 (2019-08-19)
===================
 - Change Linux release artifact back to default archive format

0.18.0 (2019-08-14)
===================
 - Add `transitions` field for State resource (#74)

0.17.0 (2019-08-06)
===================
 - Update Extension resource to add `timeout_in_ms` (#80)
 - Update ShippingMethod resource to add `predicate` (#82)

0.16.0 (2019-07-22)
===================
 - Add support for setting the externalOAuth field on the project resource (#73)
 - Add support for the StateRole Return item (#77)

0.15.1 (2019-07-16)
===================
- Trying to fix Brew release now that version number is in binary

0.15.0 (2019-07-16)
===================
- Use new Commercetools Go SDK definitions (main change is auto generated
  services, most CRUD actions are renamed)
- Fix Goreleaser not putting version number in released binary

0.14.0 (2019-07-04)
===================
- Use new Commercetools Go SDK definitions (main change is Reference is now
  ResourceIdentifier resource)

0.13.1 (2019-07-02)
===================
- Small fix for incorrect binary name in homebrew installation

0.13.0 (2019-07-02)
===================
- Add brew install option to goreleaser, see README for more info

0.12.1 (2019-06-26)
===================
- Small fix in creating api client with new scopes for resource `commercetools_api_client`

0.12.0 (2019-06-26)
===================
- *Backwards incompatible* Changed scope type from string to set for `commercetools_api_client`

0.11.1 (2019-06-26)
===================
- Fix shipping zone rate creation / deletion, thanks to @sshibani !

0.11.0 (2019-06-20)
===================
- Use new Commercetools Go SDK to set the User-Agent header on Commercetools HTTP requests.

0.10.0 (2019-06-17)
===================
 - Use Terraform 0.12.2 dependency for compatability with latest version

0.9.0 (2019-05-20)
==================
 - Use Terraform 0.12 dependency to prepare for 0.12 release

0.8.0 (2019-05-20)
==================
 - **New resource:** `commercetools_state`

0.7.1 (2019-05-14)
==================
 - Add validation for currency codes `commercetools_shipping_zone_rate`

0.7.0 (2019-05-08)
==================
 - Use latest commercetools Go SDK
 - **New resource:** `commercetools_store` **This is an alpha Commercetools resource**

0.6.0 (2019-04-26)
==================
 - **New resource:** `commercetools_shipping_method`
 - **New resource:** `commercetools_shipping_zone_rate`, *Subject to changes, tiers/validation is not yet implemented*

0.5.4 (2019-04-14)
==================
 - Fixed localized enum values being updated even if not changed for `commercetools_product_type`

0.5.3 (2019-03-27)
==================
 - Implement description update on `commercetools_product_type`
 - Implement localized enum label change on `commercetools_product_type`

0.5.2 (2019-03-26)
==================
 - Fix error reading field type `Money` in `commercetools_type`

0.5.1 (2019-03-20)
==================
 - Fix import existing `commercetools_tax_category_rate`
 - Fix tax rate edge case for 0 amount `commercetools_tax_category_rate`
 - Added docs for resource `commercetools_tax_category_rate`

0.5.0 (2019-03-19)
==================
 - **New resource** `commercetools_tax_category_rate`
 - Resource tax category: removed `rate`, now a separate resource.
 - Fix shipping zone add/remove location logic.

0.4.2 (2019-02-11)
==================
 - Fix tax rate 0.0 case not being handled

0.4.1 (2019-01-28)
==================
 - Fix resource\_type attribute label not mapping correctly
 - **New resource:** `commercetools_shipping_zone`

0.4.0 (2019-01-10)
==================
 - Use auto-generated commercetools-go-sdk types

0.3.0 (2018-12-10)
==================
 - **New resource:** `commercetools_channel`
 - **New resource:** `commercetools_tax_category`
 - Resource product type: made `attribute` elements optional
 - Resource product type: Validate/protect `required` element on Product type attribute
 - Resource type: Validate/protect `required` element on Type attribute
 - Resource type: Avoid `changeAttributeOrder` update action when new attribute gets added
 - Resource product type: Avoid `changeAttributeOrder` update action when new attribute gets added
 - Resource product type: Added support for Nested types

0.2.0 (2018-12-10)
==================
 - **New resource:** `commercetools_product_type`

0.1.1 (2018-10-04)
==================
 - **New resource:** `commercetools_type`

0.1.0 (2018-09-14)
==================
 - **New resource:** `commercetools_api_extension`
 - **New resource:** `commercetools_subscription`
 - **New resource:** `commercetools_project_settings`
