resource "commercetools_associate_role" "regional_manager" {
  key              = "regional-manager-europe"
  buyer_assignable = false
  name             = "Regional Manager - Europe"
  permissions = [
    "AddChildUnits",
    "UpdateAssociates",
    "UpdateBusinessUnitDetails",
    "UpdateParentUnit",
    "ViewMyCarts",
    "ViewOthersCarts",
    "UpdateMyCarts",
    "UpdateOthersCarts",
    "CreateMyCarts",
    "CreateOthersCarts",
    "DeleteMyCarts",
    "DeleteOthersCarts",
    "ViewMyOrders",
    "ViewOthersOrders",
    "UpdateMyOrders",
    "UpdateOthersOrders",
    "CreateMyOrdersFromMyCarts",
    "CreateMyOrdersFromMyQuotes",
    "CreateOrdersFromOthersCarts",
    "CreateOrdersFromOthersQuotes",
    "ViewMyQuotes",
    "ViewOthersQuotes",
    "AcceptMyQuotes",
    "AcceptOthersQuotes",
    "DeclineMyQuotes",
    "DeclineOthersQuotes",
    "RenegotiateMyQuotes",
    "RenegotiateOthersQuotes",
    "ReassignMyQuotes",
    "ReassignOthersQuotes",
    "ViewMyQuoteRequests",
    "ViewOthersQuoteRequests",
    "UpdateMyQuoteRequests",
    "UpdateOthersQuoteRequests",
    "CreateMyQuoteRequestsFromMyCarts",
    "CreateQuoteRequestsFromOthersCarts",
    "CreateApprovalRules",
    "UpdateApprovalRules",
    "UpdateApprovalFlows",
  ]
}
