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

## Example #1: global scoped
resource "waypoint_config_source" "globalvault" {
  type  = "globalvault"
  scope = "global"
  config = {
    addr           = "https://localhost:8200"
    skip_verify    = true
    aws_access_key = "testing"
  }
}

## Example #2: project scoped
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