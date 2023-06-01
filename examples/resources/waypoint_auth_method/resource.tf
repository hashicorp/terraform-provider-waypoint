terraform {
  required_providers {
    waypoint = {
      source  = "hashicorp/waypoint"
      version = "0.1.0"
    }
  }
}

provider "waypoint" {
  # if running locally: localhost:9701
  host = ""
  # output from `waypoint user token`
  token = ""
}

resource "waypoint_auth_method" "okta" {
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

