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
  token = "BCkP8cw7qjs4cC1Lc58KeXeLmaak4qDUUYrTs4f2R4yocugY6dymbiNhwE1SNnh4F8EFrVWY3pbk31nbmEszhyUGkG9HXViEwWMbzzd7TbSsV2RNNzhSfC6wCfnbfJQxmaugvW6HvNjHUMP2G"
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

  # project_variables = {
  #   name       = "devopsrob"
  #   job        = "dev-advocate"
  #   conference = "HashiConf EU 2022"
  # }
  project_variables = [
    {
      name  = "name"
      value = "devopsrob"
      # finalValue= null,
      sensitive = true
    },
    {
      name  = "job"
      value = "dev-advocate"
      # finalValue= null,
      sensitive = false
    },
    {
      name  = "conference"
      value = "HashiConf EU 2022"
      # finalValue= null,
      sensitive = false
    },
  ]
  # git_auth_ssh = {
  #   git_user        = "catsby"
  #   passphrase      = "test"
  #   ssh_private_key = file("~/.ssh/test-key.pem")
  # }

  git_auth_basic = {
    username = "catsby"
    password = "test"
  }
}
