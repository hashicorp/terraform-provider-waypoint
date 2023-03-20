package waypoint

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-waypoint/internal/defaults"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	Variables            []*variablesModel `tfsdk:"project_variables"`
	RemoteRunnersEnabled types.Bool        `tfsdk:"remote_runners_enabled"`
	AppStatusPollSeconds types.Int64       `tfsdk:"app_status_poll_seconds"`

	DataSourceGit *dataSourceGitModel `tfsdk:"data_source_git"`
	GitAuthBasic  *gitAuthBasicModel  `tfsdk:"git_auth_basic"`
	GitAuthSSH    *gitAuthSSHModel    `tfsdk:"git_auth_ssh"`
}

// variablesModel map variables
type variablesModel struct {
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
	Sensitive types.Bool   `tfsdk:"sensitive"`
}

// dataSourceGitModel map git information
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
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							defaults.BoolDefaultValue(types.BoolValue(false)),
						},
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
	// Get current state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectName := state.Name.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_project", projectName)

	// Get refreshed order value from HashiCups
	project, err := r.client.GetProject(ctx, projectName)
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
				gab.Password = types.StringValue(gitAuth.Basic.Password)
			case *gen.Job_Git_Ssh:
				gas = &gitAuthSSHModel{}
				gas.User = types.StringValue(gitAuth.Ssh.User)
				gas.Passphrase = types.StringValue(gitAuth.Ssh.Password)
				gas.PrivateKey = types.StringValue(string(gitAuth.Ssh.PrivateKeyPem))
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

// Delete deletes the resource and removes the Terraform state on success.
func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectName := state.Name.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_project", projectName)

	// Delete existing order
	err := r.client.DestroyProject(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Waypoint Project",
			"Could not delete project, unexpected error: "+err.Error(),
		)
		return
	}
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
	projectName := plan.Name.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_project", projectName)

	projectConf := waypointClient.DefaultProjectConfig()

	// // Git configuration for Waypoint project
	gitConfig := waypointClient.Git{
		Url:                      plan.DataSourceGit.Url.ValueString(),
		Path:                     plan.DataSourceGit.Path.ValueString(),
		IgnoreChangesOutsidePath: plan.DataSourceGit.IgnoreChangesOutsidePath.ValueBool(),
		Ref:                      plan.DataSourceGit.Ref.ValueString(),
	}

	// if len(authBasicList) > 0 {
	if plan.GitAuthBasic != nil {
		auth := &waypointClient.GitAuthBasic{
			Username: plan.GitAuthBasic.Username.ValueString(),
			Password: plan.GitAuthBasic.Password.ValueString(),
		}
		gitConfig.Auth = auth
	} else if plan.GitAuthSSH != nil {
		gitUser := plan.GitAuthSSH.User.ValueString()
		sshPrivateKey := plan.GitAuthSSH.PrivateKey.ValueString()
		var passphrase string
		if !plan.GitAuthSSH.Passphrase.IsNull() {
			passphrase = plan.GitAuthSSH.Passphrase.ValueString()
		}

		auth := &waypointClient.GitAuthSsh{
			User:          gitUser,
			PrivateKeyPem: []byte(sshPrivateKey),
			Password:      passphrase,
		}
		gitConfig.Auth = auth

	} else {
		gitConfig.Auth = nil
	}

	// // Project variables configuration
	var variableList []*gen.Variable

	for _, variable := range plan.Variables {
		projectVariable := waypointClient.SetVariable()
		projectVariable.Name = variable.Name.ValueString()
		projectVariable.Value = &gen.Variable_Str{Str: variable.Value.ValueString()}
		projectVariable.Sensitive = variable.Sensitive.ValueBool()
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
	_, err := r.client.UpsertProject(ctx, projectConf, &gitConfig, variableList)
	return plan, err
}
