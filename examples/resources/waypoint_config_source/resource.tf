resource "waypoint_config_source" "globalvault" {
  type        = "globalvault"
  scope       = "global"
  config = {
    addr           = "https://localhost:8200"
    skip_verify    = true
    aws_access_key = "testing"
  }
}

resource "waypoint_config_source" "projectvault" {
  type        = "vault"
  scope       = "project"
  project     = "test"
  application = "thing"
  config = {
    addr           = "https://localhost:8200"
    skip_verify    = false
    aws_access_key = "vault_proj"
  }
}