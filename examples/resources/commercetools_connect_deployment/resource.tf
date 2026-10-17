variable "algolia_admin_key" {
  type      = string
  sensitive = true
}

resource "commercetools_connect_deployment" "search_sync" {
  key           = "search-sync"
  connector_key = commercetools_connect_connector.search_sync.key
  region        = "europe-west1.gcp"
  type          = "production"

  # The first deployment uses the version Terraform published. Releases after
  # that are rolled out from CI/CD with the commercetools CLI, so the version
  # is ignored from then on:
  #
  #   commercetools connect connectorstaged update --key evolve-search-sync --repository-tag v1.1.0
  #   commercetools connect connectorstaged publish --key evolve-search-sync
  #   commercetools connect deployment redeploy --key search-sync --updateConnector
  connector_version = commercetools_connect_connector.search_sync.published_version

  global_standard_configuration = {
    CTP_REGION = "europe-west1.gcp"
  }

  application {
    name = "indexer"
    standard_configuration = {
      ALGOLIA_APP_ID = "MYAPPID"
    }
    secured_configuration = {
      ALGOLIA_ADMIN_KEY = var.algolia_admin_key
    }
  }

  application {
    name     = "reindex"
    schedule = "0 3 * * *"
  }

  lifecycle {
    ignore_changes = [connector_version]
  }
}
