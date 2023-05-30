package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccConfigSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: providerConfig + testAccExampleResourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("waypoint_config_source.vaultbasic", "type", "vault"),
					resource.TestCheckResourceAttr("waypoint_config_source.vaultbasic", "scope", "global"),
				),
			},
		},
	})
}

var testAccExampleResourceConfig = `
resource "waypoint_config_source" "vaultbasic" {
  type  = "vault"
  scope = "global"
  config = {
    addr           = "https://localhost:8200"
    skip_verify    = true
    aws_access_key = "testing"
  }
}
`
