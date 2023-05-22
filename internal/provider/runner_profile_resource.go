package provider

import (
	"context"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	gen "github.com/hashicorp-dev-advocates/waypoint-client/pkg/waypoint"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-provider-waypoint/internal/defaults"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &runnerProfileResource{}
	_ resource.ResourceWithConfigure = &runnerProfileResource{}
)

const defaultODRImage = "hashicorp/waypoint-odr:latest"

// NewRunnerProfileResource is a helper function to simplify the provider implementation.
func NewRunnerProfileResource() resource.Resource {
	return &runnerProfileResource{}
}

// runnerProfileResource is the resource implementation.
type runnerProfileResource struct {
	client waypointClient.Waypoint
}

// profileResourceModel maps the data schema data.
type profileResourceModel struct {
	ID                   types.String      `tfsdk:"id"`
	Name                 types.String      `tfsdk:"name"`
	OciURL               types.String      `tfsdk:"oci_url"`
	PluginType           types.String      `tfsdk:"plugin_type"`
	PluginConfig         types.String      `tfsdk:"plugin_config"`
	PluginConfigFormat   types.String      `tfsdk:"plugin_config_format"`
	Default              types.Bool        `tfsdk:"default"`
	TargetRunnerId       types.String      `tfsdk:"target_runner_id"`
	EnvironmentVariables map[string]string `tfsdk:"environment_variables"`
	TargetRunnerLabels   map[string]string `tfsdk:"target_runner_labels"`
}

// Metadata returns the resource type name.
func (r *runnerProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_runner_profile"
}

// Configure adds the provider configured client to the resource.
func (r *runnerProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(waypointClient.Waypoint)
}

// Schema defines the schema for the resource.
func (r *runnerProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Waypoint generated ID for the runner config",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the runner profile",
			},
			"oci_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "oci_url is the OCI image that will be used to boot the on demand runner.",
				PlanModifiers: []planmodifier.String{
					defaults.StringDefaultValue(types.StringValue(defaultODRImage)),
				},
			},
			"plugin_type": schema.StringAttribute{
				Required:    true,
				Description: "Plugin type for runner i.e docker / kubernetes / aws-ecs.",
			},
			"plugin_config": schema.StringAttribute{
				Optional:    true,
				Description: "Plugin config is the configuration for the plugin that is created. It is usually HCL and is decoded like the other plugins, and is plugin specific.",
			},
			"plugin_config_format": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					defaults.StringDefaultValue(types.StringValue("HCL")),
				},
				Description: "Config format specifies the format of plugin_config. Valid values are HCL or JSON. The default is HCL",
			},
			"default": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Indicates if this runner profile is the default for any new projects. The default is false",
				PlanModifiers: []planmodifier.Bool{
					defaults.BoolDefaultValue(types.BoolValue(false)),
				},
			},
			"target_runner_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the target runner for this profile.",
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("target_runner_labels"),
					}...),
				},
			},
			"target_runner_labels": schema.MapAttribute{
				Optional:    true,
				Description: "A map of labels on target runners",
				ElementType: types.StringType,
				Validators: []validator.Map{
					mapvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("target_runner_id"),
					}...),
				},
			},
			"environment_variables": schema.MapAttribute{
				Optional:    true,
				Description: "Any environment variables that should be exposed to the on demand runner.",
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *runnerProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating Runner Profile")
	// Retrieve values from plan
	var plan profileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating runner profile",
			"Could not create runner profile, unexpected error: "+err.Error(),
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
func (r *runnerProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state profileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	profileName := state.Name.ValueString()
	profileID := state.ID.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_runner_profile", profileName)

	runnerProfile, err := r.client.GetRunnerProfile(ctx, profileID)
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Info(ctx, "Runner Profile not found, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading Runner Profile",
			"Could not read runner profile with name "+profileName+": "+err.Error(),
		)
		return
	}
	// re-add the ID here, the response doesn't have it
	state.ID = types.StringValue(runnerProfile.Config.GetId())
	state.Name = types.StringValue(runnerProfile.Config.Name)
	state.OciURL = types.StringValue(runnerProfile.Config.OciUrl)
	state.PluginType = types.StringValue(runnerProfile.Config.PluginType)
	if runnerProfile.Config.GetPluginConfig() == nil {
		state.PluginConfig = types.StringNull()
	} else {
		state.PluginConfig = types.StringValue(string(runnerProfile.Config.PluginConfig))
	}
	state.PluginConfigFormat = types.StringValue(runnerProfile.Config.ConfigFormat.String())
	state.Default = types.BoolValue(runnerProfile.Config.Default)

	state.TargetRunnerId = types.StringNull()
	if runner := runnerProfile.Config.GetTargetRunner(); runner != nil {
		// It is possible this runner is empty, which can happen if the
		// configuration references a target runner id that doesn't actually
		// exist. We need to check if the ID here is nil before trying to set
		// things
		if id := runner.GetId(); id != nil {
			state.TargetRunnerId = types.StringValue(id.GetId())
		}

		if labels := runner.GetLabels(); labels != nil {
			state.TargetRunnerLabels = labels.GetLabels()
		}
	}

	state.EnvironmentVariables = runnerProfile.Config.GetEnvironmentVariables()

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *runnerProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating Runner Profile")
	// Retrieve values from plan
	var plan profileResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "waypoint_runner_profile_id", plan.ID.String())

	var err error
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating runner profile",
			"Could not update runner profile, unexpected error: "+err.Error(),
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

func (r *runnerProfileResource) upsert(ctx context.Context, plan profileResourceModel) (profileResourceModel, error) {
	profileName := plan.Name.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_runner_profile", profileName)

	runnerConfig := waypointClient.DefaultRunnerConfig()
	runnerConfig.Name = profileName

	if ociURL := plan.OciURL.ValueString(); ociURL != "" {
		runnerConfig.OciUrl = ociURL
	}

	if pluginType := plan.PluginType.ValueString(); pluginType != "" {
		runnerConfig.PluginType = pluginType
	}

	if pluginConfig := plan.PluginConfig.ValueString(); pluginConfig != "" {
		runnerConfig.PluginConfig = []byte(pluginConfig)
	}

	if pluginConfigFormat := plan.PluginConfigFormat.ValueString(); pluginConfigFormat != "" {
		switch pluginConfigFormat {
		case "HCL":
			// HCL is 0
			runnerConfig.ConfigFormat = 0
		case "JSON":
			// JSON is 1
			runnerConfig.ConfigFormat = 1
		default:
			// error
		}
	}

	if defaultProfile := plan.Default.ValueBool(); defaultProfile {
		runnerConfig.Default = defaultProfile
	}

	if !plan.TargetRunnerId.IsNull() && plan.TargetRunnerId.ValueString() != "" {
		runnerConfig.TargetRunner = &gen.Ref_Runner{
			Target: &gen.Ref_Runner_Id{
				Id: &gen.Ref_RunnerId{
					Id: plan.TargetRunnerId.ValueString(),
				},
			},
		}
	}
	tRL := plan.TargetRunnerLabels

	if len(tRL) > 0 {
		runnerConfig.TargetRunner.Target = &gen.Ref_Runner_Labels{
			Labels: &gen.Ref_RunnerLabels{
				Labels: tRL,
			}}
	}

	if environmentVariables := plan.EnvironmentVariables; environmentVariables != nil {
		runnerConfig.EnvironmentVariables = environmentVariables
	}

	// Upsert the profile; the method CreateRunnerProfile itself uses upsert
	runnerProfile, err := r.client.CreateRunnerProfile(ctx, runnerConfig)
	plan.ID = types.StringValue(runnerProfile.Config.GetId())

	return plan, err
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *runnerProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state profileResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	profileID := state.ID.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_runner_profile", profileID)

	// Delete existing profile
	err := r.client.DeleteRunnerProfile(ctx, profileID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Waypoint Runner Profile",
			"Could not delete runner profile, unexpected error: "+err.Error(),
		)
		return
	}
}
