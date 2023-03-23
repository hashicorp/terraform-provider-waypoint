package provider

import (
	"context"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &runnerProfileDataSource{}
	_ datasource.DataSourceWithConfigure = &runnerProfileDataSource{}
)

// NewRunnerProfileDataSource is a helper function to simplify the provider implementation.
func NewRunnerProfileDataSource() datasource.DataSource {
	return &runnerProfileDataSource{}
}

// runnerProfileDataSource is the data source implementation.
type runnerProfileDataSource struct {
	client waypointClient.Waypoint
}

// Configure adds the provider configured client to the data source.
func (d *runnerProfileDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(waypointClient.Waypoint)
}

// Metadata returns the data source type name.
func (d *runnerProfileDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner_profile"
}

// profileDataModel maps the data schema data.
type profileDataModel struct {
	ID                   types.String      `tfsdk:"id"`
	Name                 types.String      `tfsdk:"profile_name"`
	OciURL               types.String      `tfsdk:"oci_url"`
	PluginType           types.String      `tfsdk:"plugin_type"`
	PluginConfig         types.String      `tfsdk:"plugin_config"`
	TargetRunnerId       types.String      `tfsdk:"target_runner_id"`
	PluginConfigFormat   types.String      `tfsdk:"plugin_config_format"`
	Default              types.Bool        `tfsdk:"default"`
	EnvironmentVariables map[string]string `tfsdk:"environment_variables"`
	TargetRunnerLabels   map[string]string `tfsdk:"target_runner_labels"`
	// EnvironmentVariables types.Map `tfsdk:"environment_variables"`
}

// Schema defines the schema for the data source.
func (d *runnerProfileDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Schema defines the schema for the data source.
		Attributes: map[string]schema.Attribute{
			"profile_name": schema.StringAttribute{
				// Optional:    true,
				Computed:    true,
				Description: "The name of the Runner profile",
			},
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The id of the Runner profile",
			},
			"oci_url": schema.StringAttribute{
				Computed:    true,
				Description: "oci_url is the OCI image that will be used to boot the on demand runner.",
			},
			"plugin_type": schema.StringAttribute{
				Computed:    true,
				Description: "Plugin type for runner i.e docker / kubernetes / aws-ecs.",
			},
			"plugin_config": schema.StringAttribute{
				Computed:    true,
				Description: "plugin config is the configuration for the plugin that is created. It is usually HCL and is decoded like the other plugins, and is plugin specific.",
			},
			"target_runner_id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the target runner for this profile.",
			},
			"default": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates if this runner profile is the default for any new projects",
			},
			"plugin_config_format": schema.StringAttribute{
				Computed:    true,
				Description: "config format specifies the format of plugin_config.",
			},
			"environment_variables": schema.MapAttribute{
				Computed:    true,
				Description: "Any env vars that should be exposed to the on demand runner.",
				ElementType: types.StringType,
			},
			"target_runner_labels": schema.MapAttribute{
				Computed:    true,
				Description: "A map of labels on target runners",
				ElementType: types.StringType,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *runnerProfileDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read refreshes the Terraform state with the latest data.
	var state profileDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	getRunnerProfile, err := d.client.GetRunnerProfile(context.TODO(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Runner Profile",
			"Could not read Runner Profile with ID"+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}
	profile := getRunnerProfile.Config
	state.Name = types.StringValue(profile.GetName())
	state.OciURL = types.StringValue(profile.GetOciUrl())
	state.PluginType = types.StringValue(profile.GetPluginType())
	state.PluginConfig = types.StringValue(string(profile.GetPluginConfig()))

	// Target Runner here is either an ID or a list of labels
	if targetRunner := profile.GetTargetRunner(); targetRunner != nil {
		if id := targetRunner.GetId(); id != nil {
			state.TargetRunnerId = types.StringValue(id.GetId())
		}
		if labelsRaw := targetRunner.GetLabels(); labelsRaw != nil {
			state.TargetRunnerLabels = labelsRaw.GetLabels()
		}
	}

	state.Default = types.BoolValue(profile.GetDefault())
	state.PluginConfigFormat = types.StringValue(profile.GetConfigFormat().String())
	state.EnvironmentVariables = profile.EnvironmentVariables

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
