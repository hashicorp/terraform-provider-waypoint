package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	tfConfig, err := helperTestAccTFExampleConfig("resources/waypoint_project/resource.tf")
	if err != nil {
		t.Errorf("error reading config from file: %s", err)
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: tfConfig,
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

					// Example1 from examples directory
					resource.TestCheckResourceAttr("waypoint_project.example1", "project_name", "example1"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "remote_runners_enabled", "true"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "data_source_git.git_url", "https://github.com/hashicorp/waypoint-examples"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "data_source_git.git_path", "docker/go"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "data_source_git.git_ref", "HEAD"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "data_source_git.file_change_signal", "some-signal"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "data_source_git.git_poll_interval_seconds", "15"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "app_status_poll_seconds", "12"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "project_variables.0.name", "devopsrob"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "project_variables.0.value", "dev-advocate"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "project_variables.0.sensitive", "false"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "git_auth_ssh.%", "3"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "git_auth_ssh.git_user", "cassie"),
					resource.TestCheckResourceAttr("waypoint_project.example1", "git_auth_ssh.passphrase", "test"),
				),
				ExpectError: nil,
				PlanOnly:    false,
			},
		},
	})
}
