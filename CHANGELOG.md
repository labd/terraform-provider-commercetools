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
