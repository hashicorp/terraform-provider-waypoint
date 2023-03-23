package provider

import (
	"context"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework-validators/mapvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &runnerProfileResource{}
	_ resource.ResourceWithConfigure = &runnerProfileResource{}
)

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
	Name                 types.String      `tfsdk:"profile_name"`
	OciURL               types.String      `tfsdk:"oci_url"`
	PluginType           types.String      `tfsdk:"plugin_type"`
	PluginConfig         types.String      `tfsdk:"plugin_config"`
	TargetRunnerId       types.String      `tfsdk:"target_runner_id"`
	PluginConfigFormat   types.String      `tfsdk:"plugin_config_format"`
	Default              types.Bool        `tfsdk:"default"`
	EnvironmentVariables map[string]string `tfsdk:"environment_variables"`
	TargetRunnerLabels   map[string]string `tfsdk:"target_runner_labels"`
}

// Metadata returns the resource type name.
func (r *runnerProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "runner_profile"
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
			},
			"profile_name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the runner profile",
			},
			"oci_url": schema.StringAttribute{
				Optional:    true,
				Description: "oci_url is the OCI image that will be used to boot the on demand runner.",
			},
			"plugin_type": schema.StringAttribute{
				Optional:    true,
				Description: "Plugin type for runner i.e docker / kubernetes / aws-ecs.",
			},
			"plugin_config": schema.StringAttribute{
				Optional:    true,
				Description: "plugin config is the configuration for the plugin that is created. It is usually HCL and is decoded like the other plugins, and is plugin specific.",
			},
			"plugin_config_format": schema.Int64Attribute{
				Optional:    true,
				Description: "config format specifies the format of plugin_config.",
			},
			"default": schema.BoolAttribute{
				Optional:    true,
				Description: "Indicates if this runner profile is the default for any new projects",
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
				Description: "Any env vars that should be exposed to the on demand runner.",
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *runnerProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
}

// Read refreshes the Terraform state with the latest data.
func (r *runnerProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *runnerProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *runnerProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
