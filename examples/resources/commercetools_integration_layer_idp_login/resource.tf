variable "idp_client_secret" {
  type      = string
  sensitive = true
}

resource "commercetools_integration_layer_idp_login" "login" {
  issuer              = "https://login.example.com"
  token_endpoint      = "https://login.example.com/oauth2/token"
  revocation_endpoint = "https://login.example.com/oauth2/revoke"
  jwks_uri            = "https://login.example.com/.well-known/jwks.json"
  client_id           = "storefront"
  client_secret       = var.idp_client_secret
  redirect_uri        = "https://shop.example.com/auth/callback"
  match_strategy      = "email"
}
