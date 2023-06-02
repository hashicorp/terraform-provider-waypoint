// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
)

var (
	_ provider.Provider = &waypointProvider{}
)

type waypointProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

type waypointProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// New creates a new WaypointProvider
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &waypointProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name.
func (p *waypointProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "waypoint"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *waypointProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"token": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (p *waypointProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring waypoint client")
	// Retrieve provider data from configuration
	var config waypointProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown waypoint API Host",
			"The provider cannot create the waypoint API client as there is an unknown configuration value for the waypoint API host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the WAYPOINT_HOST environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown Waypoint API token",
			"The provider cannot create the Waypoint API client as there is an unknown configuration value for the waypoint API token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the WAYPOINT_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	host := os.Getenv("WAYPOINT_HOST")
	token := os.Getenv("WAYPOINT_TOKEN")

	if host == "" {
		host = config.Host.ValueString()
	}

	if token == "" {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing waypoint API Host",
			"The provider cannot create the waypoint API client as there is a missing or empty value for the waypoint API host. "+
				"Set the host value in the configuration or use the WAYPOINT_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing Waypoint API token",
			"The provider cannot create the waypoint API client as there is a missing or empty value for the waypoint API token. "+
				"Set the token value in the configuration or use the WAYPOINT_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "waypoint_host", host)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "waypoint_token")

	if token == "" {
		resp.Diagnostics.AddError(
			"no token provided",
			"A Waypoint token is required to use this provider",
		)
		return
	}

	tflog.Debug(ctx, "Creating waypoint client")
	waypointClientConfig := waypointClient.DefaultConfig()
	waypointClientConfig.Address = host
	waypointClientConfig.Token = token

	wc, err := waypointClient.New(waypointClientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create waypoint API Client",
			"An unexpected error occurred when creating the waypoint API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"waypoint Client Error: "+err.Error(),
		)
		return
	}

	// Make the waypoint client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = wc
	resp.ResourceData = wc

	tflog.Info(ctx, "Configured waypoint client", map[string]any{"success": true})
}

// DataSources defines the data sources implemented in the provider.
func (p *waypointProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewAuthMethodDataSource,
		NewProjectDataSource,
		NewRunnerProfileDataSource,
		NewAppDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *waypointProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewAuthMethodResource,
		NewConfigSourceResource,
		NewProjectResource,
		NewRunnerProfileResource,
	}
}
