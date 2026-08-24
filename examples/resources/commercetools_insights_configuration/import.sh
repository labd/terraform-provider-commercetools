# The Platform Insights configuration is a singleton per Project, so it is
# imported using the Project key.
#
# Provider credentials are not returned by the API: api_key and OTLP header
# values are empty after an import and need one apply to be set again.
terraform import commercetools_insights_configuration.my-insights-configuration my-project-key
