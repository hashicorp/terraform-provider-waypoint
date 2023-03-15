package waypoint

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	gen "github.com/hashicorp-dev-advocates/waypoint-client/pkg/waypoint"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &projectResource{}
	_ resource.ResourceWithConfigure = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client waypointClient.Waypoint
}

// projectResourceModel maps the resource schema data.
type projectResourceModel struct {
	Name                 types.String      `tfsdk:"project_name"`
	Variables            map[string]string `tfsdk:"project_variables"`
	RemoteRunnersEnabled types.Bool        `tfsdk:"remote_runners_enabled"`
	AppStatusPollSeconds types.Int64       `tfsdk:"app_status_poll_seconds"`

	DataSourceGit *dataSourceGitModel `tfsdk:"data_source_git"`
	GitAuthBasic  *gitAuthBasicModel  `tfsdk:"git_auth_basic"`
	GitAuthSSH    *gitAuthSSHModel    `tfsdk:"git_auth_ssh"`
}

// dataSourceGitModel maps data source data
type dataSourceGitModel struct {
	Url                      types.String `tfsdk:"git_url"`
	Path                     types.String `tfsdk:"git_path"`
	Ref                      types.String `tfsdk:"git_ref"`
	IgnoreChangesOutsidePath types.Bool   `tfsdk:"ignore_changes_outside_path"`
	PollInterval             types.Int64  `tfsdk:"git_poll_interval_seconds"`
	FileChangeSignal         types.String `tfsdk:"file_change_signal"`
}

// gitAuthBasicModel maps git auth data
type gitAuthBasicModel struct {
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
}

// gitAuthSSHModel maps git auth ssh data
type gitAuthSSHModel struct {
	User       types.String `tfsdk:"git_user"`
	Passphrase types.String `tfsdk:"passphrase"`
	PrivateKey types.String `tfsdk:"ssh_private_key"`
}

// Configure adds the provider configured client to the resource.
func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(waypointClient.Waypoint)
}

// Schema defines the schema for the resource.
func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Waypoint project",
			},
			"project_variables": schema.MapAttribute{
				Optional:    true,
				Description: "List of variables in Key/value pairs associated with the Waypoint Project",
				ElementType: types.StringType,
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
				Optional:  true,
				Sensitive: true,
				Validators: []validator.Object{
					objectvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("git_auth_basic"),
					}...),
				},
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

// Create creates the resource and sets the initial Terraform state.
func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating Project")
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project",
			"Could not create project, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// 	client := m.(*WaypointClient).conn

	// 	projectName := d.Get("project_name").(string)
	// 	project, err := client.GetProject(context.TODO(), projectName)
	// 	if err != nil {
	// 		return diag.Errorf("Error retrieving the %s project", projectName)
	// 	}
	// Get current state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get refreshed order value from HashiCups
	project, err := r.client.GetProject(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Project",
			"Could not read Project with name "+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	// 	d.Set("remote_runners_enabled", project.RemoteEnabled)
	state.RemoteRunnersEnabled = types.BoolValue(project.RemoteEnabled)

	// 	applications := flattenApplications(project.Applications)
	// 	d.Set("applications", applications)

	// 	variables := flattenVariables(project.Variables)
	// 	d.Set("project_variables", variables)

	// 	dataSourceGitSlice := map[string]interface{}{}
	// 	dataSourceGitSlice["git_url"] = project.DataSource.GetGit().Url
	// 	dataSourceGitSlice["git_path"] = project.DataSource.GetGit().Path
	// 	dataSourceGitSlice["git_ref"] = project.DataSource.GetGit().Ref
	// 	dataSourceGitSlice["file_change_signal"] = project.FileChangeSignal

	// 	dpi, _ := time.ParseDuration(project.DataSourcePoll.Interval)
	// 	dataSourceGitSlice["git_poll_interval_seconds"] = dpi / time.Second
	// 	d.Set("data_source_git", []interface{}{dataSourceGitSlice})

	// 	gitAuthBasicSlice := map[string]interface{}{}
	// 	gitAuthSshSlice := map[string]interface{}{}

	// 	gitAuth := project.DataSource.Source.(*gen.Job_DataSource_Git).Git.Auth
	// 	switch gitAuth.(type) {
	// 	case *gen.Job_Git_Basic_:
	// 		gitAuthBasicSlice["username"] = gitAuth.(*gen.Job_Git_Basic_).Basic.Username
	// 		gitAuthBasicSlice["password"] = gitAuth.(*gen.Job_Git_Basic_).Basic.Password
	// 		d.Set("git_auth_basic", []interface{}{gitAuthBasicSlice})
	// 	case *gen.Job_Git_Ssh:
	// 		gitAuthSshSlice["git_user"] = gitAuth.(*gen.Job_Git_Ssh).Ssh.User
	// 		gitAuthSshSlice["passphrase"] = gitAuth.(*gen.Job_Git_Ssh).Ssh.Password
	// 		gitAuthSshSlice["ssh_private_key"] = string(gitAuth.(*gen.Job_Git_Ssh).Ssh.PrivateKeyPem)
	// 		d.Set("git_auth_ssh", []interface{}{gitAuthSshSlice})
	// 	}

	// 	if project.StatusReportPoll != nil {
	// 		asps := project.StatusReportPoll.Interval
	// 		aspsParse, _ := time.ParseDuration(asps)
	// 		d.Set("app_status_poll_seconds", aspsParse/time.Second)
	// 	}

	// return nil
	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

// Metadata returns the resource type name.
func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating Project")
	// Retrieve values from plan
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating project",
			"Could not update project, unexpected error: "+err.Error(),
		)
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) upsert(ctx context.Context, plan projectResourceModel) (projectResourceModel, error) {
	projectConf := waypointClient.DefaultProjectConfig()

	// // Git configuration for Waypoint project
	var gitConfig *waypointClient.Git

	// if len(authBasicList) > 0 {
	if plan.GitAuthBasic != nil {
		// q.Q("found git auth basic")
		// var auth *client.GitAuthBasic

		// authBasicSlice := authBasicList[0].(map[string]interface{})
		// username := authBasicSlice["username"]
		// password := authBasicSlice["password"]

		// auth = &client.GitAuthBasic{
		// 	Username: username.(string),
		// 	Password: password.(string),
		// }

		// gitConfig = &client.Git{
		// 	Url:                      dataSourceSlice["git_url"].(string),
		// 	Path:                     dataSourceSlice["git_path"].(string),
		// 	IgnoreChangesOutsidePath: dataSourceSlice["ignore_changes_outside_path"].(bool),
		// 	Ref:                      dataSourceSlice["git_ref"].(string),
		// 	Auth:                     auth,
		// }
	} else if plan.GitAuthSSH != nil {
		// q.Q("found git auth ssh")
		// 	var auth *client.GitAuthSsh
		// 	authSshSlice := authSshList[0].(map[string]interface{})
		// 	var passphrase interface{}
		// 	gitUser := authSshSlice["git_user"]
		// 	sshPrivateKey := authSshSlice["ssh_private_key"]
		// 	if authSshSlice["passphrase"] != nil {
		// 		passphrase = authSshSlice["passphrase"]
		// 	}

		// 	auth = &client.GitAuthSsh{
		// 		User:          gitUser.(string),
		// 		PrivateKeyPem: []byte(sshPrivateKey.(string)),
		// 		Password:      passphrase.(string),
		// 	}

		// 	gitConfig = &client.Git{
		// 		Url:                      dataSourceSlice["git_url"].(string),
		// 		Path:                     dataSourceSlice["git_path"].(string),
		// 		IgnoreChangesOutsidePath: dataSourceSlice["ignore_changes_outside_path"].(bool),
		// 		Ref:                      dataSourceSlice["git_ref"].(string),
		// 		Auth:                     auth,
		// 	}

	} else {
		// q.Q("found basic git stuff")
		gitConfig = &waypointClient.Git{
			Url:                      plan.DataSourceGit.Url.ValueString(),
			Path:                     plan.DataSourceGit.Path.ValueString(),
			IgnoreChangesOutsidePath: plan.DataSourceGit.IgnoreChangesOutsidePath.ValueBool(),
			Ref:                      plan.DataSourceGit.Ref.ValueString(),
			Auth:                     nil,
		}
	}

	// // Project variables configuration
	var variableList []*gen.Variable
	// varsList := d.Get("project_variables").(map[string]interface{})

	for key, value := range plan.Variables {
		projectVariable := waypointClient.SetVariable()
		projectVariable.Name = key
		projectVariable.Value = &gen.Variable_Str{Str: value}
		variableList = append(variableList, &projectVariable)
	}

	// Project config for request
	projectConf.Name = plan.Name.ValueString()
	projectConf.RemoteRunnersEnabled = plan.RemoteRunnersEnabled.ValueBool()

	// TODO: verify this is the correct value being sent if value is not
	// specified, defaulted, etc.
	projectConf.StatusReportPoll = time.Duration(plan.AppStatusPollSeconds.ValueInt64()) * time.Second
	projectConf.GitPollInterval = time.Duration(plan.DataSourceGit.PollInterval.ValueInt64()) * time.Second

	projectConf.FileChangeSignal = plan.DataSourceGit.FileChangeSignal.ValueString()

	// q.Q("project:", projectConf)
	// q.Q("git config:", gitConfig)
	// q.Q("variableList:", variableList)

	_, err := r.client.UpsertProject(ctx, projectConf, gitConfig, variableList)
	return plan, err
}
