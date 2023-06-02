terraform {
  required_providers {
    waypoint = {
      source = "hashicorp/waypoint"
      # version = ""
      # latest version by default
      # see the following resources for more information on specific versions:
      # https://github.com/hashicorp/terraform-provider-waypoint/blob/main/CHANGELOG.md
      # https://releases.hashicorp.com/
      # https://github.com/hashicorp/terraform-provider-waypoint/releases
    }
  }
}

provider "waypoint" {
  # if running locally: localhost:9701
  host = ""
  # output from `waypoint user token`
  token = ""
}