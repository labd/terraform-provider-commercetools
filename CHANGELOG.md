Unreleased
==========
 - Resource API Extension: Removed unused `azure_functions` type

v0.27.0 (2021-03-01)
====================
 - Resource project: Add `carts` field with countryTaxRateFallBackEnabled setting 
 - Resource project: Fix updating of `messages` field to explicitly set `false` when deleted or set to false in terraform instead of relying on commercetools default settings for project in these scenarios 

v0.26.1 (2021-01-21)
====================
 - Resource api_extension: Fixed typo in `trigger` field name that caused updates to actions in triggers to fail

v0.26.0 (2021-01-12)
====================
 - **New resource** `commercetools_customer_group` (#141)
 - Resource type: Allow updating the label of an existing Enum value
 - Resource type: Add support to update a set of enum in a custom type
 - Fix ProductType and DiscountCode tests with real commercetools environment

v0.25.3 (2020-12-17)
====================
 - Resource store: Force creation of new resource when changing the keyL there is no update action for this available.

v0.25.2 (2020-12-17)
====================
 - Resource channel: Add support for updating the `key` field

v0.25.1 (2020-12-05)
====================
 - Resource type: Fix a bug when the `input_hint` of a field was modified.

v0.25.0 (2020-11-27)
====================
 - **New resource:** `commercetools_custom_object`

v0.24.1 (2020-11-13)
====================
 - Resource tax_rate: Add a workaround to handle an issue with changing id's after a tax category update

v0.24.0 (2020-11-13)
====================
 - Resource store: Add `supply_channels` field
 - Resource tax_category_rate: Handle non-existing tax rates when refreshing state

v0.23.0 (2020-08-28)
====================
 - New tag naming scheme (prefix with v) to conform to terraform repository requirements
 - Update terraform-plugin-sdk for compatibility with terraform 0.13

0.22.1 (2020-07-21)
===================
- Resource store: Fix edge case where distribution channels were not updated

0.22.0 (2020-07-20)
===================
- Resource store: Add `distribution_channels` field
- Update commercetools-go-sdk dependency to v0.2.0. This version now properly
  handles oauth2 authentication failures (#117)

0.21.2 (2020-06-11)
===================
- Resource store: Add `languages` field

0.21.1 (2020-04-22)
===================
- Resource channel: Fix read null pointer exception. Name should be optional.

0.21.0 (2020-02-27)
===================
- Provider arguments (`client_id`, `client_secret`, `project_key`,
  `scopes`, `token_url` and `api_url`) are now required
- Resource api_client: Updating now recreates the resource since
  it cannot be updated.
- Don't retry various calls if Commercetools returns an error (resulting in
  unnecessary retries/waiting times).
- Dependency update: use terraform-plugin-sdk 1.7.0

0.20.0 (2020-02-22)
===================
- Resource subscription: Add Azure Event Grid support

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
 - Resource state: Add `transitions` field (#74)

0.17.0 (2019-08-06)
===================
 - Resource api_extension: Update Extension resource to add `timeout_in_ms` (#80)
 - Resource shipping_method: Add `predicate` field (#82)

0.16.0 (2019-07-22)
===================
 - Resource project: Add support for setting the externalOAuth field (#73)
 - Resource state: Add support for the StateRole Return item (#77)

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
- Resource api_client: Small fix in creating api client with new scopes

0.12.0 (2019-06-26)
===================
**Breaking chanages**

- Resource api_client: Changed scope type from string to set

0.11.1 (2019-06-26)
===================
- Resource shipping_zone: Fix creation and deletion, thanks to @sshibani !

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
 - Resource shipping_zone_rate: Add validation for currency codes

0.7.0 (2019-05-08)
==================
 - **New resource:** `commercetools_store` **This is an alpha Commercetools resource**
 - Use latest commercetools Go SDK

0.6.0 (2019-04-26)
==================
 - **New resource:** `commercetools_shipping_method`
 - **New resource:** `commercetools_shipping_zone_rate`, *Subject to changes, tiers/validation is not yet implemented*

0.5.4 (2019-04-14)
==================
 - Resource product_type: Fixed localized enum values being updated even if not changed

0.5.3 (2019-03-27)
==================
 - Resource product_type: Implement description update
 - Resource product_type: Implement localized enum label change

0.5.2 (2019-03-26)
==================
 - Resource type: Fix error reading field type `Money`

0.5.1 (2019-03-20)
==================
 - Resource tax_category_rate: Fix import existing instance
 - Resource tax_category_rate: Fix tax rate edge case for 0 amount

0.5.0 (2019-03-19)
==================
 - **New resource** `commercetools_tax_category_rate`
 - Resource tax_category: removed `rate` in favour of `commercetools_tax_category_rate`
 - Resource shipping_zone: Fix add/remove location logic.

0.4.2 (2019-02-11)
==================
 - Resource tax_category: Fix tax rate 0.0 case not being handled

0.4.1 (2019-01-28)
==================
 - **New resource:** `commercetools_shipping_zone`
 - Fix resource\_type attribute label not mapping correctly

0.4.0 (2019-01-10)
==================
 - Use auto-generated commercetools-go-sdk types

0.3.0 (2018-12-10)
==================
 - **New resource:** `commercetools_channel`
 - **New resource:** `commercetools_tax_category`
 - Resource product_type: made `attribute` elements optional
 - Resource product_type: Validate/protect `required` element on Product type attribute
 - Resource product_type: Avoid `changeAttributeOrder` update action when new attribute gets added
 - Resource product_type: Added support for Nested types
 - Resource type: Validate/protect `required` element on Type attribute
 - Resource type: Avoid `changeAttributeOrder` update action when new attribute gets added

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
