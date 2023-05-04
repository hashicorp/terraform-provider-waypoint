package provider

import (
	"context"
	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &appDataSource{}
	_ datasource.DataSourceWithConfigure = &appDataSource{}
)

// NewAppDataSource is a helper function to simplify the provider implementation.
func NewAppDataSource() datasource.DataSource {
	return &appDataSource{}
}

// appDataSource is the data source implementation.
type appDataSource struct {
	client waypointClient.Waypoint
}

// appDataSource maps the schema data.
type appDataSourceModel struct {
	Name             types.String `tfsdk:"app_name"`
	Project          types.String `tfsdk:"project_name"`
	FileChangeSignal types.String `tfsdk:"file_change_signal"`
}

// Configure adds the provider configured client to the data source.
func (d *appDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(waypointClient.Waypoint)
}

// Metadata returns the data source type name.
func (d *appDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

// Schema defines the schema for the data source
func (d *appDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"app_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Waypoint application.",
			},
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Waypoint project.",
			},
			"file_change_signal": &schema.StringAttribute{
				Computed:    true,
				Description: "Indicates signal to be sent to any applications when their config files change.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *appDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var state appDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := state.Name.ValueString()
	projName := state.Project.ValueString()
	ctx = tflog.SetField(ctx, "application", appName)

	// Get app based on tf config
	app, err := d.client.GetApp(ctx, appName, projName)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Info(ctx, "Application not found, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Application",
			"Could not read Application with name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// Set var back to state
	state.Project = types.StringValue(app.Project.Project)
	state.Name = types.StringValue(app.Name)
	state.FileChangeSignal = types.StringValue(app.FileChangeSignal)

	// Set refreshed state to see if there is a diff btw plans
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
