terraform {
  required_providers {
    waypoint = {
      source  = "hashicorp/waypoint"
      version = "0.1.0"
    }
  }
}

provider "waypoint" {
  waypoint_addr = "localhost:9701"
}

data "waypoint_application" "example" {
  app_name     = "app"
  project_name = "proj"
}

