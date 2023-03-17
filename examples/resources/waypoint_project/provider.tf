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

resource "waypoint_project" "example" {

  project_name           = "example"
  remote_runners_enabled = true

  data_source_git = {
    git_url                   = "https://github.com/hashicorp/waypoint-examples"
    git_path                  = "docker/go"
    git_ref                   = "HEAD"
    file_change_signal        = "some-signal"
    git_poll_interval_seconds = 15
    # ignore_changes_outside_path = true
  }

  app_status_poll_seconds = 12

  project_variables = [
    {
      name      = "name"
      value     = "devopsrob"
      sensitive = true
    },
    {
      name      = "job"
      value     = "dev-advocate"
      sensitive = false
    },
    {
      name      = "conference"
      value     = "HashiConf EU 2022"
      sensitive = false
    },
  ]

  git_auth_basic = {
    username = "catsby"
    password = "test"
  }
}
