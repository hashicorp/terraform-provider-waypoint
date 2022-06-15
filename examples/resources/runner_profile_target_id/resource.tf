resource "waypoint_runner_profile" "target_id" {
  profile_name     = "summer"
  oci_url          = "hashicorp/waypoint-odr:latest"
  plugin_type      = "docker"
  default          = true
  target_runner_id = "01G5GNJEYC7RVJNXFGMHD0HCDT"

  environment_variables = {
    VAULT_ADDR           = "https://localhost:8200"
    VAULT_CLIENT_TIMEOUT = "30s"
  }
}

