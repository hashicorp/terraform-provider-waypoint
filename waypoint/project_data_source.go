package waypoint

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource = &projectDataSource{}
)

// NewProjectDataSource is a helper function to simplify the provider implementation.
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

// projectDataSource is the data source implementation.
type projectDataSource struct{}

// Metadata returns the data source type name.
func (d *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema defines the schema for the data source
func (d *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Waypoint project",
			},
			"project_variables": schema.ListNestedAttribute{
				Optional:    true,
				Description: "List of variables in Key/value pairs associated with the Waypoint Project",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true},
						"value": schema.StringAttribute{
							Required: true},
						"sensitive": schema.BoolAttribute{
							Optional: true,
						},
					},
				},
			},
			"data_source_git": &schema.SingleNestedAttribute{
				Required:    true,
				Description: "Configuration of Git repository where waypoint.hcl file is stored",
				Attributes: map[string]schema.Attribute{
					"git_url": &schema.StringAttribute{
						Optional:    true,
						Description: "Url of git repository storing the waypoint.hcl file",
					},
					"git_path": &schema.StringAttribute{
						Optional:    true,
						Description: "Path in git repository when waypoint.hcl file is stored in a sub-directory",
					},
					"git_ref": &schema.StringAttribute{
						Optional:    true,
						Description: "Git repository ref containing waypoint.hcl file",
					},
					"ignore_changes_outside_path": &schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Description: "Whether Waypoint ignores changes outside path storing waypoint.hcl file",
					},
					"git_poll_interval_seconds": &schema.Int64Attribute{
						Optional:    true,
						Description: "Interval at which Waypoint should poll git repository for changes",
					},
					"file_change_signal": &schema.StringAttribute{
						Optional:    true,
						Description: "Indicates signal to be sent to any applications when their config files change.",
					},
				},
			},
			"remote_runners_enabled": &schema.BoolAttribute{
				Optional:    true,
				Description: "Enable remote runners for project",
			},
			"git_auth_basic": &schema.SingleNestedAttribute{
				Optional:    true,
				Description: "Basic authentication details for Git consisting of `username` and `password`",
				Sensitive:   true,
				Attributes: map[string]schema.Attribute{
					"username": &schema.StringAttribute{
						Required:    true,
						Description: "Git username",
					},
					"password": &schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "Git password",
					},
				},
			},
			"git_auth_ssh": &schema.SingleNestedAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "SSH authentication details for Git",
				Attributes: map[string]schema.Attribute{
					"git_user": &schema.StringAttribute{
						Optional:    true,
						Description: "Git user associated with private key",
					},
					"passphrase": &schema.StringAttribute{
						Optional:    true,
						Sensitive:   true,
						Description: "Passphrase to use with private key",
					},
					"ssh_private_key": &schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "Private key to authenticate to Git",
					},
				},
			},
			"app_status_poll_seconds": &schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Application status poll interval in seconds",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
}
