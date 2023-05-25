# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

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

data "waypoint_runner_profile" "kube" {
  id = "01GVRNP5XG2SEYYA564CS6BDDJ"
}
data "waypoint_runner_profile" "nomad_labels" {
  id = "01GVRVP869G887K7XWP595A47H"
}
data "waypoint_runner_profile" "nomad_id" {
  id = "01GVRVQ5Z95K1ZN1K74GYR4V2X"
}


output "default_profile" {
  value = data.waypoint_runner_profile.default_docker
}

output "kube_profile" {
  value = data.waypoint_runner_profile.kube
}

output "nomad_labels" {
  value = data.waypoint_runner_profile.nomad_labels
}

output "nomad_id" {
  value = data.waypoint_runner_profile.nomad_id
}
