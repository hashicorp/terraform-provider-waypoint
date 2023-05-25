# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

data "waypoint_app" "example" {
  app_name     = "example-nodejs"
  project_name = "example-nodejs"
}