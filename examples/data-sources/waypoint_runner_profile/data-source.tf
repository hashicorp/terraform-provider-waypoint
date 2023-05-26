terraform {
  required_providers {
    waypoint = {
      source  = "hashicorp/waypoint"
      version = "0.1.0"
    }
  }
}

provider "waypoint" {
  host  = "localhost:9701"
  token = ""
}

data "waypoint_runner_profile" "default_docker" {
  id = "01GV45AW59XGNT906S8XXKG5E5"
}

output "default_profile" {
  value = data.waypoint_runner_profile.default_docker
}
