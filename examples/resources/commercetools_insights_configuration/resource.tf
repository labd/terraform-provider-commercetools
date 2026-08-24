resource "commercetools_insights_configuration" "my-insights-configuration" {
  active = true

  apm_provider {
    type        = "NewRelic"
    event_types = ["Logs", "Metrics"]
    region      = "eu"
    api_key     = var.newrelic_api_key
  }

  apm_provider {
    type        = "Datadog"
    event_types = ["Metrics"]
    site        = "EU1"
    api_key     = var.datadog_api_key
  }

  apm_provider {
    type        = "OtlpHttp"
    event_types = ["Metrics"]
    endpoint    = "https://my-otlp-provider.example.com:14318"

    header {
      name  = "api-key"
      value = var.otlp_api_key
    }
  }

  additional_attribute {
    event_type = "Metrics"
    name       = "team"
    value      = "commercetools"
  }
}
