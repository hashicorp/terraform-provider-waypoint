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

data "waypoint_project" "example" {
  project_name = "example"
}
