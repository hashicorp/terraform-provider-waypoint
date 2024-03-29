---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "waypoint_project Resource - terraform-provider-waypoint"
subcategory: ""
description: |-
  
---

# waypoint_project (Resource)



## Example Usage

```terraform
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

##Git auth ssh example
resource "waypoint_project" "example1" {

  project_name           = "example1"
  remote_runners_enabled = true

  data_source_git = {
    git_url                   = "https://github.com/hashicorp/waypoint-examples"
    git_path                  = "docker/go"
    git_ref                   = "HEAD"
    file_change_signal        = "some-signal"
    git_poll_interval_seconds = 15
  }

  app_status_poll_seconds = 12

  project_variables = [
    {
      name      = "devopsrob"
      value     = "dev-advocate"
      sensitive = "false"
    },
  ]

  git_auth_ssh = {
    git_user        = "cassie"
    passphrase      = "test"
    ssh_private_key = <<EOF
-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCjcGqTkOq0CR3rTx0ZSQSIdTrDrFAYl29611xN8aVgMQIWtDB/
lD0W5TpKPuU9iaiG/sSn/VYt6EzN7Sr332jj7cyl2WrrHI6ujRswNy4HojMuqtfa
b5FFDpRmCuvl35fge18OvoQTJELhhJ1EvJ5KUeZiuJ3u3YyMnxxXzLuKbQIDAQAB
AoGAPrNDz7TKtaLBvaIuMaMXgBopHyQd3jFKbT/tg2Fu5kYm3PrnmCoQfZYXFKCo
ZUFIS/G1FBVWWGpD/MQ9tbYZkKpwuH+t2rGndMnLXiTC296/s9uix7gsjnT4Naci
5N6EN9pVUBwQmGrYUTHFc58ThtelSiPARX7LSU2ibtJSv8ECQQDWBRrrAYmbCUN7
ra0DFT6SppaDtvvuKtb+mUeKbg0B8U4y4wCIK5GH8EyQSwUWcXnNBO05rlUPbifs
DLv/u82lAkEAw39sTJ0KmJJyaChqvqAJ8guulKlgucQJ0Et9ppZyet9iVwNKX/aW
9UlwGBMQdafQ36nd1QMEA8AbAw4D+hw/KQJBANJbHDUGQtk2hrSmZNoV5HXB9Uiq
7v4N71k5ER8XwgM5yVGs2tX8dMM3RhnBEtQXXs9LW1uJZSOQcv7JGXNnhN0CQBZe
nzrJAWxh3XtznHtBfsHWelyCYRIAj4rpCHCmaGUM6IjCVKFUawOYKp5mmAyObkUZ
f8ue87emJLEdynC1CLkCQHduNjP1hemAGWrd6v8BHhE3kKtcK6KHsPvJR5dOfzbd
HAqVePERhISfN6cwZt5p8B3/JUwSR8el66DF7Jm57BM=
-----END RSA PRIVATE KEY-----
EOF
  }
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `data_source_git` (Attributes) Configuration of Git repository where waypoint.hcl file is stored (see [below for nested schema](#nestedatt--data_source_git))
- `project_name` (String) The name of the Waypoint project

### Optional

- `app_status_poll_seconds` (Number) Application status poll interval in seconds
- `git_auth_basic` (Attributes, Sensitive) Basic authentication details for Git consisting of `username` and `password` (see [below for nested schema](#nestedatt--git_auth_basic))
- `git_auth_ssh` (Attributes, Sensitive) SSH authentication details for Git (see [below for nested schema](#nestedatt--git_auth_ssh))
- `project_variables` (Attributes List) List of variables in Key/value pairs associated with the Waypoint Project (see [below for nested schema](#nestedatt--project_variables))
- `remote_runners_enabled` (Boolean) Enable remote runners for project

### Read-Only

- `id` (String) The id required for acceptance testing to work

<a id="nestedatt--data_source_git"></a>
### Nested Schema for `data_source_git`

Optional:

- `file_change_signal` (String) Indicates signal to be sent to any applications when their config files change.
- `git_path` (String) Path in git repository when waypoint.hcl file is stored in a sub-directory
- `git_poll_interval_seconds` (Number) Interval at which Waypoint should poll git repository for changes
- `git_ref` (String) Git repository ref containing waypoint.hcl file
- `git_url` (String) Url of git repository storing the waypoint.hcl file
- `ignore_changes_outside_path` (Boolean) Whether Waypoint ignores changes outside path storing waypoint.hcl file


<a id="nestedatt--git_auth_basic"></a>
### Nested Schema for `git_auth_basic`

Required:

- `password` (String, Sensitive) Git password
- `username` (String) Git username


<a id="nestedatt--git_auth_ssh"></a>
### Nested Schema for `git_auth_ssh`

Required:

- `ssh_private_key` (String, Sensitive) Private key to authenticate to Git

Optional:

- `git_user` (String) Git user associated with private key
- `passphrase` (String, Sensitive) Passphrase to use with private key


<a id="nestedatt--project_variables"></a>
### Nested Schema for `project_variables`

Required:

- `name` (String)
- `value` (String)

Optional:

- `sensitive` (Boolean)


