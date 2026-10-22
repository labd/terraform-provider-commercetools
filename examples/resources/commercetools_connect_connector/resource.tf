resource "commercetools_connect_connector" "search_sync" {
  key               = "evolve-search-sync"
  name              = "Evolve search sync"
  description       = "Keeps the Algolia index in sync with the catalog."
  repository_url    = "git@github.com:labd/evolve-search-sync.git"
  repository_tag    = "v1.0.0"
  integration_types = ["search"]
  supported_regions = ["europe-west1.gcp"]
  private_projects  = ["europe-west1.gcp:my-project"]
  publish           = true

  creator {
    email   = "dev@example.com"
    company = "Example Company"
  }

  lifecycle {
    # Releases set a new tag with the commercetools CLI.
    ignore_changes = [repository_tag]
  }
}
