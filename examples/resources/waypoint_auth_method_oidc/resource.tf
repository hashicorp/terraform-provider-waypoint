# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "waypoint_auth_method_oidc" "okta" {
  name          = "my-oidc"
  display_name  = "My OIDC Provider"
  client_id     = "..."
  client_secret = "..."
  discovery_url = "https://my-oidc.provider/oauth2/default"
  allowed_redirect_urls = [
    "https://localhost:9702/auth/oidc-callback",
  ]

  auds = [
    "..."
  ]

  list_claim_mappings = {
    groups = "groups"
  }

  signing_algs = [
    "rsa512"
  ]

  discovery_ca_pem = [
    "cert1.crt"
  ]
}