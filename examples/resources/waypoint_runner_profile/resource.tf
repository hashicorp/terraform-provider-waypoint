## Example with runner labels
resource "waypoint_runner_profile" "target_id" {
  profile_name = "example"
  oci_url      = "hashicorp/waypoint-odr:latest"
  plugin_type  = "docker"
  default      = true
  target_runner_labels = {
    app = "payments"
  }

  environment_variables = {
    VAULT_ADDR           = "https://localhost:8200"
    VAULT_CLIENT_TIMEOUT = "30s"
  }
}

## Example with runner id
resource "waypoint_runner_profile" "example" {
  profile_name     = "example"
  oci_url          = "hashicorp/waypoint-odr:latest"
  plugin_type      = "docker"
  default          = true
  target_runner_id = "01G5GNJEYC7RVJNXFGMHD0HCDT"

  environment_variables = {
    VAULT_ADDR           = "https://localhost:8200"
    VAULT_CLIENT_TIMEOUT = "30s"
  }
}
