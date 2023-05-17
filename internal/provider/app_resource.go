package provider

import (
	"context"
	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	gen "github.com/hashicorp-dev-advocates/waypoint-client/pkg/waypoint"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/pkg/errors"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &appResource{}
	_ resource.ResourceWithConfigure = &appResource{}
)

func NewAppResource() resource.Resource {
	return &appResource{}
}

type appResource struct {
	client waypointClient.Waypoint
}

// appResourceModel maps the resource schema data.
type appResourceModel struct {
	Name             types.String `tfsdk:"app_name"`
	Project          types.String `tfsdk:"project_name"`
	FileChangeSignal types.String `tfsdk:"file_change_signal"`
}

// Configure adds the provider configured client to the resource.
func (r *appResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(waypointClient.Waypoint)
}

// Schema defines the schema for the resource.
func (r *appResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Optional:    true,
				Description: "Indicates signal to be sent to any applications when their config files change.",
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *appResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating App")
	// Retrieve values from plan
	var plan appResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	if plan.Name.ValueString() == "" || plan.Project.ValueString() == "" {
		resp.Diagnostics.AddError(
			"App and Project are both needed for app lookup",
			"Please ensure that you have both an app and project defined in terraform.",
		)
		return
	}
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating app",
			"Could not create app, unexpected error: "+err.Error(),
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
func (r *appResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, "Reading App")

	//Get current state
	var state appResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := state.Name.ValueString()
	projName := state.Project.ValueString()

	ctx = tflog.SetField(ctx, "application", appName)

	// Get app based on tf config
	app, err := r.client.GetApp(ctx, appName, projName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading App",
			"Could not find App with name: "+state.Name.ValueString()+" and project: "+state.Project.ValueString()+". "+err.Error(),
		)

		return
	}

	// Set vars back to state
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

// Delete deletes the resource and removes the Terraform state on success.
func (r *appResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "Deleting App")

	// Retrieve values from state
	var state appResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.AddError("App deletion is currently unimplemented due to this logic not existing in Waypoint", "UNIMPLEMENTED")

	//appName := state.Name.ValueString()
	//ctx = tflog.SetField(ctx, "waypoint_app", appName)

	// Delete existing application
	//err := r.client.DestroyApp(ctx, state.Name.ValueString())
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"Error Deleting Waypoint App",
	//		"Could not delete app, unexpected error: "+err.Error(),
	//	)
	//	return
	//}
}

// Metadata returns the resource type name.
func (r *appResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_app"
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *appResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating App")

	// Retrieve values from plan
	var plan appResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var err error
	plan, err = r.upsert(ctx, plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating app",
			"Could not update app, unexpected error: "+err.Error(),
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

func (r *appResource) upsert(ctx context.Context, plan appResourceModel) (appResourceModel, error) {
	appName := plan.Name.ValueString()

	ctx = tflog.SetField(ctx, "waypoint_app", appName)

	appConf := waypointClient.DefaultApplicationConfig()
	if plan.Name.ValueString() == "" || plan.Project.ValueString() == "" {
		return plan, errors.New("Please ensure that you have both an app and project defined in terraform")
	}
	// App config for request
	appConf.Name = appName
	appConf.Project = &gen.Ref_Project{Project: plan.Project.ValueString()}
	appConf.FileChangeSignal = plan.FileChangeSignal.ValueString()

	_, err := r.client.UpsertApp(ctx, appConf)

	return plan, err
}
