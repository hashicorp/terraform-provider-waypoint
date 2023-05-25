// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"time"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	gen "github.com/hashicorp-dev-advocates/waypoint-client/pkg/waypoint"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &projectDataSource{}
	_ datasource.DataSourceWithConfigure = &projectDataSource{}
)

// NewProjectDataSource is a helper function to simplify the provider implementation.
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

// projectDataSource is the data source implementation.
type projectDataSource struct {
	client waypointClient.Waypoint
}

// projectDataSourceModel maps the schema data. This embeds the
// projectResourceModel struct, and adds Application information
type projectDataSourceModel struct {
	Applications         types.List        `tfsdk:"applications"`
	Name                 types.String      `tfsdk:"project_name"`
	Variables            []*variablesModel `tfsdk:"project_variables"`
	RemoteRunnersEnabled types.Bool        `tfsdk:"remote_runners_enabled"`
	AppStatusPollSeconds types.Int64       `tfsdk:"app_status_poll_seconds"`

	DataSourceGit *dataSourceGitModel `tfsdk:"data_source_git"`
	GitAuthBasic  *gitAuthBasicModel  `tfsdk:"git_auth_basic"`
	GitAuthSSH    *gitAuthSSHModel    `tfsdk:"git_auth_ssh"`
}

// Configure adds the provider configured client to the data source.
func (d *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(waypointClient.Waypoint)
}

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
			"applications": schema.ListAttribute{
				Computed:    true,
				Description: "List of applications for this project",
				ElementType: types.StringType,
			},
			"project_variables": schema.ListNestedAttribute{
				Computed:    true,
				Sensitive:   true,
				Description: "List of variables in Key/value pairs associated with the Waypoint Project",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required: true},
						"value": schema.StringAttribute{
							Required: true},
						"sensitive": schema.BoolAttribute{
							Computed: true,
						},
					},
				},
			},
			"data_source_git": &schema.SingleNestedAttribute{
				Computed:    true,
				Description: "Configuration of Git repository where waypoint.hcl file is stored",
				Attributes: map[string]schema.Attribute{
					"git_url": &schema.StringAttribute{
						Computed:    true,
						Description: "Url of git repository storing the waypoint.hcl file",
					},
					"git_path": &schema.StringAttribute{
						Computed:    true,
						Description: "Path in git repository when waypoint.hcl file is stored in a sub-directory",
					},
					"git_ref": &schema.StringAttribute{
						Computed:    true,
						Description: "Git repository ref containing waypoint.hcl file",
					},
					"ignore_changes_outside_path": &schema.BoolAttribute{
						Computed:    true,
						Description: "Whether Waypoint ignores changes outside path storing waypoint.hcl file",
					},
					"git_poll_interval_seconds": &schema.Int64Attribute{
						Computed:    true,
						Description: "Interval at which Waypoint should poll git repository for changes",
					},
					"file_change_signal": &schema.StringAttribute{
						Computed:    true,
						Description: "Indicates signal to be sent to any applications when their config files change.",
					},
				},
			},
			"remote_runners_enabled": &schema.BoolAttribute{
				Computed:    true,
				Description: "Enable remote runners for project",
			},
			"git_auth_basic": &schema.SingleNestedAttribute{
				Computed:    true,
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
				Computed:    true,
				Sensitive:   true,
				Description: "SSH authentication details for Git",
				Attributes: map[string]schema.Attribute{
					"git_user": &schema.StringAttribute{
						Computed:    true,
						Description: "Git user associated with private key",
					},
					"passphrase": &schema.StringAttribute{
						Computed:    true,
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
				Computed:    true,
				Description: "Application status poll interval in seconds",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Get current state
	var state projectDataSourceModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectName := state.Name.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_project", projectName)

	// Get refreshed order value from HashiCups
	project, err := d.client.GetProject(ctx, projectName)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Info(ctx, "Project not found, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Project",
			"Could not read Project with name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// build up list of applications
	var appListStr []string
	if apps := project.GetApplications(); apps != nil {
		for _, app := range apps {
			appListStr = append(appListStr, app.GetName())
		}
	}

	apps, diags := types.ListValueFrom(ctx, types.StringType, appListStr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Applications = apps

	state.RemoteRunnersEnabled = types.BoolValue(project.RemoteEnabled)

	var projectVariables []*variablesModel
	for _, v := range project.Variables {
		pvar := variablesModel{}
		pvar.Name = types.StringValue(v.Name)
		// we only support string values(?)
		str := v.Value.(*gen.Variable_Str).Str
		pvar.Value = types.StringValue(str)
		pvar.Sensitive = types.BoolValue(v.Sensitive)
		projectVariables = append(projectVariables, &pvar)
	}
	state.Variables = projectVariables

	// TODO: This should be a &gen.Job_DataSource_Git from the protos. Not yet
	// sure what to do here if it's a _Local or _Remote version of that
	// dataSource := project.DataSource.Source
	var dsg *dataSourceGitModel
	var gab *gitAuthBasicModel
	var gas *gitAuthSSHModel
	if project.DataSource != nil {
		switch project.DataSource.Source.(type) {
		case *gen.Job_DataSource_Local, *gen.Job_DataSource_Remote:
		// not sure what to do here
		default:
			// assumes *gen.Job_DataSource_Git
			src := project.DataSource.Source.(*gen.Job_DataSource_Git)
			pollRaw, _ := time.ParseDuration(project.DataSourcePoll.Interval)
			poll := pollRaw / time.Second
			dsg = &dataSourceGitModel{
				Url:                      types.StringValue(src.Git.Url),
				Ref:                      types.StringValue(src.Git.Ref),
				Path:                     types.StringValue(src.Git.Path),
				IgnoreChangesOutsidePath: types.BoolValue(src.Git.IgnoreChangesOutsidePath),
				PollInterval:             types.Int64Value(int64(poll)),
				FileChangeSignal:         types.StringValue(project.FileChangeSignal),
			}

			authRaw := src.Git.Auth
			switch gitAuth := authRaw.(type) {
			case *gen.Job_Git_Basic_:
				gab = &gitAuthBasicModel{}
				gab.Username = types.StringValue(gitAuth.Basic.Username)
			case *gen.Job_Git_Ssh:
				gas = &gitAuthSSHModel{}
				gas.User = types.StringValue(gitAuth.Ssh.User)
			}
		}
	}
	state.DataSourceGit = dsg
	state.GitAuthBasic = gab
	state.GitAuthSSH = gas

	if project.StatusReportPoll != nil {
		pollRaw, _ := time.ParseDuration(project.StatusReportPoll.Interval)
		poll := pollRaw / time.Second
		state.AppStatusPollSeconds = types.Int64Value(int64(poll))
	}

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
