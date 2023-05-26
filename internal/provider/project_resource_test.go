package provider

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"testing"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				//PreConfig: func() {
				//	testAccPreCheck(t)
				//},
				Config: providerConfig + `
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
`,
				//ResourceName: "waypoint_project.example",
				//PreConfig: nil,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("waypoint_project.example", "project_name", "example"),
					resource.TestCheckResourceAttr("waypoint_project.example", "remote_runners_enabled", "true"),
					resource.TestCheckResourceAttr("waypoint_project.example", "data_source_git.git_url", "https://github.com/hashicorp/waypoint-examples"),
					resource.TestCheckResourceAttr("waypoint_project.example", "app_status_poll_seconds", "12"),
					//resource.TestCheckResourceAttr("waypoint_project.example", "project_variables", "__"),
					//resource.TestCheckResourceAttr("waypoint_project.example", "git_auth_basic", "__"),
				),
				//Destroy: true,
				//ExpectNonEmptyPlan: true,
				ExpectError:        nil,
				PlanOnly:           false,
				PreventDiskCleanup: false, //false for cleanup, should be true long term
				//PreventPostDestroyRefresh: false,
				//ImportState:               false,
				//RefreshState:              false,
			},
		},
	})
}

func testAccProjectConfig(filename string) string {
	fileContent, err := os.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
	}

	return string(fileContent)
}
