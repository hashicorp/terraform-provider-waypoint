---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "waypoint_project Data Source - terraform-provider-waypoint"
subcategory: ""
description: |-
  
---

# waypoint_project (Data Source)



## Example Usage

```terraform
data "waypoint_project" "example" {
  project_name = "example"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `project_name` (String) The name of the Waypoint project

### Read-Only

- `app_status_poll_seconds` (Number) Application status poll interval in seconds
- `applications` (List of String) List of applications for this project
- `data_source_git` (Attributes) Configuration of Git repository where waypoint.hcl file is stored (see [below for nested schema](#nestedatt--data_source_git))
- `git_auth_basic` (Attributes, Sensitive) Basic authentication details for Git consisting of `username` and `password` (see [below for nested schema](#nestedatt--git_auth_basic))
- `git_auth_ssh` (Attributes, Sensitive) SSH authentication details for Git (see [below for nested schema](#nestedatt--git_auth_ssh))
- `project_variables` (Attributes List, Sensitive) List of variables in Key/value pairs associated with the Waypoint Project (see [below for nested schema](#nestedatt--project_variables))
- `remote_runners_enabled` (Boolean) Enable remote runners for project

<a id="nestedatt--data_source_git"></a>
### Nested Schema for `data_source_git`

Read-Only:

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

Read-Only:

- `git_user` (String) Git user associated with private key
- `passphrase` (String, Sensitive) Passphrase to use with private key


<a id="nestedatt--project_variables"></a>
### Nested Schema for `project_variables`

Required:

- `name` (String)
- `value` (String)

Read-Only:

- `sensitive` (Boolean)


