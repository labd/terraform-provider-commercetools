resource "commercetools_associate_role" "regional_manager" {
     key = "regional-manager-europe"
     buyer_assignable = false
     name = "Regional Manager - Europe"
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
