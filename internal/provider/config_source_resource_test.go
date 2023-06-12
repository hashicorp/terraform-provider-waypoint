// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

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
				Config: helperTestAccTFExampleConfig(t, "resources/waypoint_config_source/resource.tf"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Example #1: global scoped
					resource.TestCheckResourceAttr("waypoint_config_source.globalvault", "type", "globalvault"),
					resource.TestCheckResourceAttr("waypoint_config_source.globalvault", "scope", "global"),

					// Example #2: project scoped
					resource.TestCheckResourceAttr("waypoint_config_source.projectvault", "type", "vault"),
					resource.TestCheckResourceAttr("waypoint_config_source.projectvault", "scope", "project"),
				),
				ExpectError: nil,
				PlanOnly:    false,
			},
		},
	})
}
