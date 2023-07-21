![Boundary Logo](Waypoint_PrimaryLogo_Color_RGB.png)

# Terraform Provider Waypoint

Available in the [Terraform Registry.](https://registry.terraform.io/providers/hashicorp/waypoint)

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.17

## Building The Provider

1. Clone the repository
1. Enter the repository directory
1. Build the provider using the make `install` command: 
```sh
$ make install
```

## Adding Dependencies

This provider uses [Go modules](https://github.com/golang/go/wiki/Modules).
Please see the Go documentation for the most up to date information about using Go modules.

To add a new dependency `github.com/author/dependency` to your Terraform provider:

```
go get github.com/author/dependency
go mod tidy
```

Then commit the changes to `go.mod` and `go.sum`.

## Using the provider

```hcl
terraform {
  required_providers {
    waypoint = {
      source  = "hashicorp/waypoint"
      version = "0.1.0"
      # latest version by default
      # see the following resources for more information on specific versions:
      # https://github.com/hashicorp/terraform-provider-waypoint/blob/main/CHANGELOG.md
      # https://releases.hashicorp.com/
      # https://github.com/hashicorp/terraform-provider-waypoint/releases
    }
  }
}

provider "waypoint" {
  # if running locally: localhost:9701, 
  # or use WAYPOINT_HOST environment variable 
  # host = ""
  
  # output from `waypoint user token`, 
  # or use WAYPOINT_TOKEN environment variable 
  # token = ""
}

resource "waypoint_project" "example" {

  project_name           = "example"
  remote_runners_enabled = true

  data_source_git {
    git_url                   = "https://github.com/hashicorp/waypoint-examples"
    git_path                  = "docker/go"
    git_ref                   = "HEAD"
    file_change_signal        = "some-signal"
    git_poll_interval_seconds = 15
  }

  app_status_poll_seconds = 12

  project_variables = {
    name       = "devopsrob"
    job        = "dev-advocate"
    conference = "HashiConf EU 2022"
  }

  git_auth_basic {
    username = "test"
    password = "test"
  }
}
```
NOTE: `WAYPOINT_HOST` and `WAYPOINT_TOKEN` may be set as environment variables, and will override the provider host/token set in your terraform configuration.


## Developing the Provider

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (see [Requirements](#requirements) above).

To compile the provider, run `go install`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

To generate or update documentation, run `go generate`.

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
