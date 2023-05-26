package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: providerConfig + testAccProjectResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("waypoint_project.example", "project_name", "example"),
					resource.TestCheckResourceAttr("waypoint_project.example", "remote_runners_enabled", "true"),
					resource.TestCheckResourceAttr("waypoint_project.example", "data_source_git.git_url", "https://github.com/hashicorp/waypoint-examples"),
					resource.TestCheckResourceAttr("waypoint_project.example", "data_source_git.git_path", "docker/go"),
					resource.TestCheckResourceAttr("waypoint_project.example", "data_source_git.git_ref", "HEAD"),
					resource.TestCheckResourceAttr("waypoint_project.example", "data_source_git.file_change_signal", "some-signal"),
					resource.TestCheckResourceAttr("waypoint_project.example", "data_source_git.git_poll_interval_seconds", "15"),
					resource.TestCheckResourceAttr("waypoint_project.example", "app_status_poll_seconds", "12"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.0.name", "name"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.0.value", "devopsrob"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.0.sensitive", "true"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.1.name", "job"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.1.value", "dev-advocate"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.1.sensitive", "false"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.2.name", "conference"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.2.value", "HashiConf EU 2022"),
					resource.TestCheckResourceAttr("waypoint_project.example", "project_variables.2.sensitive", "false"),
					resource.TestCheckResourceAttr("waypoint_project.example", "git_auth_basic.%", "2"),
					resource.TestCheckResourceAttr("waypoint_project.example", "git_auth_basic.username", "catsby"),
					resource.TestCheckResourceAttr("waypoint_project.example", "git_auth_basic.password", "test"),
				),
				ExpectError: nil,
				PlanOnly:    false,
			},
		},
	})
}

var testAccProjectResourceConfig = `
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
`
