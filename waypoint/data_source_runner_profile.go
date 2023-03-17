package waypoint

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
	Name types.String `tfsdk:"profile_name"`
	ID   types.String `tfsdk:"id"`
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
		},
		// 		Schema: map[string]*schema.Schema{
		// 			"profile_name": {
		// 				Type:        schema.TypeString,
		// 				Computed:    true,
		// 				Description: "The name of the runner profile",
		// 			},
		// 			"id": {
		// 				Type:        schema.TypeString,
		// 				Required:    true,
		// 				Description: "Computed ID of runner profile.",
		// 			},
		// 			"oci_url": {
		// 				Type:        schema.TypeString,
		// 				Computed:    true,
		// 				Description: "oci_url is the OCI image that will be used to boot the on demand runner.",
		// 			},
		// 			"plugin_type": {
		// 				Type:        schema.TypeString,
		// 				Computed:    true,
		// 				Description: "Plugin type for runner i.e docker / kubernetes / aws-ecs.",
		// 			},
		// 			"plugin_config": {
		// 				Type:        schema.TypeString, // Under the hood the type is []byte
		// 				Computed:    true,
		// 				Description: "plugin config is the configuration for the plugin that is created. It is usually HCL and is decoded like the other plugins, and is plugin specific.",
		// 			},
		// 			"plugin_config_format": {
		// 				Type:        schema.TypeInt,
		// 				Computed:    true,
		// 				Description: "config format specifies the format of plugin_config.",
		// 			},
		// 			"default": {
		// 				Type:        schema.TypeBool,
		// 				Computed:    true,
		// 				Description: "Indicates if this runner profile is the default for any new projects",
		// 			},
		// 			"target_runner_id": {
		// 				Type:        schema.TypeString,
		// 				Computed:    true,
		// 				Description: "The ID of the target runner for this profile.",
		// 			},
		// 			"target_runner_labels": {
		// 				Type:        schema.TypeMap,
		// 				Computed:    true,
		// 				Description: "A map of labels on target runners",
		// 				Elem: &schema.Schema{
		// 					Type: schema.TypeString,
		// 				},
		// 			},
		// 			"environment_variables": {
		// 				Type:        schema.TypeMap,
		// 				Computed:    true,
		// 				Description: "Any env vars that should be exposed to the on demand runner.",
		// 				Elem: &schema.Schema{
		// 					Type: schema.TypeString,
		// 				},
		// 			},
		// 		},
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
	state.Name = types.StringValue(getRunnerProfile.Config.GetName())

	// coffees, err := d.client.GetCoffees()
	// if err != nil {
	// 	resp.Diagnostics.AddError(
	// 		"Unable to Read HashiCups Coffees",
	// 		err.Error(),
	// 	)
	// 	return
	// }

	// Map response body to model
	// for _, coffee := range coffees {
	//     coffeeState := coffeesModel{
	//         ID:          types.Int64Value(int64(coffee.ID)),
	//         Name:        types.StringValue(coffee.Name),
	//         Teaser:      types.StringValue(coffee.Teaser),
	//         Description: types.StringValue(coffee.Description),
	//         Price:       types.Float64Value(coffee.Price),
	//         Image:       types.StringValue(coffee.Image),
	//     }

	//     for _, ingredient := range coffee.Ingredient {
	//         coffeeState.Ingredients = append(coffeeState.Ingredients, coffeesIngredientsModel{
	//             ID: types.Int64Value(int64(ingredient.ID)),
	//         })
	//     }

	//     state.Coffees = append(state.Coffees, coffeeState)
	// }

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}
