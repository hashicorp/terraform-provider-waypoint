package provider

import (
	"context"
	"strings"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                   = &configSourceResource{}
	_ resource.ResourceWithConfigure      = &configSourceResource{}
	_ resource.ResourceWithValidateConfig = &configSourceResource{}
)

// NewConfigSourceResource is a helper function to simplify the provider implementation.
func NewConfigSourceResource() resource.Resource {
	return &configSourceResource{}
}

// configSourceResource is the resource implementation.
type configSourceResource struct {
	client waypointClient.Waypoint
}

// configSourceResourceModel maps the data schema data.
type configSourceResourceModel struct {
	Type        types.String      `tfsdk:"type"`
	Scope       types.String      `tfsdk:"scope"`
	Project     types.String      `tfsdk:"project"`
	Application types.String      `tfsdk:"application"`
	Workspace   types.String      `tfsdk:"workspace"`
	Config      map[string]string `tfsdk:"config"`
}

// Metadata returns the resource type name.
func (r *configSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config_source"
}

// Configure adds the provider configured client to the resource.
func (r *configSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(waypointClient.Waypoint)
}

// Schema defines the schema for the resource.
func (r *configSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"type": schema.StringAttribute{
				Required:    true,
				Description: "Config Source type",
			},
			"project": schema.StringAttribute{
				Optional:    true,
				Description: "Config Source Project",
			},
			"application": schema.StringAttribute{
				Optional:    true,
				Description: "Config Source Project",
			},
			"workspace": schema.StringAttribute{
				Optional:    true,
				Description: "Config Source Workspace",
			},
			"scope": schema.StringAttribute{
				Required:    true,
				Description: "Config Source Scope",
			},
			"config": schema.MapAttribute{
				Optional:    true,
				Description: "Configuration for the dynamic source type",
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *configSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating config source")
	// Retrieve values from plan
	var plan configSourceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating config source",
			"Could not create config source, unexpected error: "+err.Error(),
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
func (r *configSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state configSourceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sourceConfig := waypointClient.DefaultConfigSourceConfig()
	sourceConfig.SourceType = state.Type.ValueString()

	ctx = tflog.SetField(ctx, "waypoint_config_source", sourceConfig.SourceType)

	// We don't have a unique ID we can reference, so we need to reload our
	// config source from state and query the endpoint for it.
	sourceConfig.Workspace = state.Workspace.ValueString()
	sourceConfig.Project = state.Project.ValueString()
	sourceConfig.Application = state.Application.ValueString()
	sourceConfig.Scope = state.Scope.ValueString()

	cfg, err := r.client.GetConfigSource(ctx, sourceConfig)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Info(ctx, "config source not found, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading config source",
			"Could not read config source with type "+sourceConfig.SourceType+": "+err.Error(),
		)
		return
	}

	ws := cfg.GetWorkspace()
	if ws == nil {
		// state.Workspace = types.StringNull()
	} else if ws.GetWorkspace() == "" {
		state.Workspace = types.StringNull()
	} else {
		state.Workspace = types.StringValue(ws.GetWorkspace())
	}

	// if runnerProfile.Config.GetPluginConfig() == nil {
	// 	state.PluginConfig = types.StringNull()
	// } else {
	// 	state.PluginConfig = types.StringValue(string(runnerProfile.Config.PluginConfig))
	// }
	// state.PluginConfigFormat = types.StringValue(runnerProfile.Config.ConfigFormat.String())
	// state.Default = types.BoolValue(runnerProfile.Config.Default)

	state.Config = cfg.GetConfig()

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *configSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating config source")
	// Retrieve values from plan
	var plan configSourceResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "waypoint_config_source_type", plan.Type.String())

	var err error
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating config source",
			"Could not update config source, unexpected error: "+err.Error(),
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

func (r *configSourceResource) upsert(ctx context.Context, plan configSourceResourceModel) (configSourceResourceModel, error) {
	sourceType := plan.Type.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_config_source", sourceType)
	sourceConfig := waypointClient.DefaultConfigSourceConfig()
	sourceConfig.SourceType = plan.Type.ValueString()
	sourceConfig.Workspace = plan.Workspace.ValueString()
	sourceConfig.Project = plan.Project.ValueString()
	sourceConfig.Application = plan.Application.ValueString()
	sourceConfig.Scope = plan.Scope.ValueString()
	if configVars := plan.Config; configVars != nil {
		sourceConfig.Config = configVars
	}

	// Upsert the config source; the method SetConfigSource itself uses upsert
	err := r.client.SetConfigSource(ctx, sourceConfig)
	if err != nil {
		return plan, err
	}

	return plan, nil
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *configSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state configSourceResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	sourceConfig := waypointClient.DefaultConfigSourceConfig()
	sourceConfig.SourceType = state.Type.ValueString()
	sourceConfig.Workspace = state.Workspace.ValueString()
	sourceConfig.Project = state.Project.ValueString()
	sourceConfig.Application = state.Application.ValueString()
	sourceConfig.Scope = state.Scope.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_config_source", sourceConfig.SourceType)

	// Delete existing config source
	err := r.client.DeleteConfigSource(ctx, sourceConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Waypoint config source",
			"Could not delete config source, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *configSourceResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var data configSourceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If scope is app, make sure app and project are set
	if strings.ToLower(data.Scope.ValueString()) == "app" {
		if data.Project.IsNull() || data.Application.IsNull() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("application"),
				"Missing Attribute Configuration",
				"Expected Project and Application to be configured when using Scope 'app'. "+
					"The resource may return unexpected results.",
			)
			return
		}
	}

	// If scope is project, make sure project is configured
	if strings.ToLower(data.Scope.ValueString()) == "project" {
		if data.Project.IsNull() {
			resp.Diagnostics.AddAttributeWarning(
				path.Root("project"),
				"Missing Attribute Configuration",
				"Expected Project to be configured when Scope is 'project'. "+
					"The resource may return unexpected results.",
			)
			return
		}
	}
}

func sourceFromState() {
}
