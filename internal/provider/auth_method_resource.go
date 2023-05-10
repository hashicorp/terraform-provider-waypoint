package provider

import (
	"context"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &authMethodResource{}
	_ resource.ResourceWithConfigure = &authMethodResource{}
)

// NewAuthMethodResource is a helper function to simplify the provider implementation.
func NewAuthMethodResource() resource.Resource {
	return &authMethodResource{}
}

// authMethodResource is the data source implementation.
type authMethodResource struct {
	client waypointClient.Waypoint
}

// Configure adds the provider configured client to the data source.
func (r *authMethodResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(waypointClient.Waypoint)
}

// Metadata returns the data source type name.
func (r *authMethodResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_method"
}

// authMethodResourceModel maps the data schema data.
type authMethodResourceModel struct {
	Name                types.String `tfsdk:"name"`
	DisplayName         types.String `tfsdk:"display_name"`
	Description         types.String `tfsdk:"description"`
	AccessorSelector    types.String `tfsdk:"accessor_selector"`
	ClientID            types.String `tfsdk:"client_id"`
	ClientSecret        types.String `tfsdk:"client_secret"`
	DiscoveryURL        types.String `tfsdk:"discovery_url"`
	AllowedRedirectURIs types.List   `tfsdk:"allowed_redirect_uris"`
	ClaimMappings       types.Map    `tfsdk:"claim_mappings"`
	ListClaimMappings   types.Map    `tfsdk:"list_claim_mappings"`
	DiscoveryCAPEM      types.List   `tfsdk:"discovery_ca_pem"`
	SigningAlgs         types.List   `tfsdk:"signing_algs"`
	Scopes              types.List   `tfsdk:"scopes"`
	Auds                types.List   `tfsdk:"auds"`
}

// Schema defines the schema for the resource.
func (r *authMethodResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Schema defines the schema for the resource.
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Auth Method",
			},
			"display_name": schema.StringAttribute{
				Optional:    true,
				Description: "The display name of the Auth Method",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of auth method",
			},
			"accessor_selector": schema.StringAttribute{
				Optional: true,
			},
			"client_id": schema.StringAttribute{
				Required:    true,
				Description: "Client ID of OIDC provider",
			},
			"client_secret": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Client Secret of OIDC provider",
			},
			"discovery_url": schema.StringAttribute{
				Required:    true,
				Description: "Discovery URL for OIDC provider",
			},
			"allowed_redirect_uris": schema.ListAttribute{
				Optional:    true,
				Description: "Allowed URI for auth redirection.",
				ElementType: types.StringType,
			},
			"claim_mappings": schema.MapAttribute{
				Optional:    true,
				Description: "Mapping of a claim to a variable value for the access selector",
				ElementType: types.StringType,
			},
			"list_claim_mappings": schema.MapAttribute{
				Optional:    true,
				Description: "Same as claim_mappings but for list values",
				ElementType: types.StringType,
			},
			"discovery_ca_pem": schema.ListAttribute{
				Optional:    true,
				Description: "Optional CA certificate chain to validate the discovery URL. Multiple CA certificates can be specified to support easier rotation",
				ElementType: types.StringType,
			},
			"signing_algs": schema.ListAttribute{
				Optional:    true,
				Description: "The signing algorithms supported by the OIDC connect server. If this isn't specified, this will default to RS256 since that should be supported according to the RFC. The string values here should be valid OIDC signing algorithms",
				ElementType: types.StringType,
			},
			"scopes": schema.ListAttribute{
				Optional:    true,
				Description: "The optional claims scope requested.",
				ElementType: types.StringType,
			},
			"auds": schema.ListAttribute{
				Optional:    true,
				Description: "The optional audience claims required",
				ElementType: types.StringType,
			},
		},
	}
}

// Create creates the resource and sets the initial Terraform state.
func (r *authMethodResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	tflog.Info(ctx, "Creating Auth Method")
	// Retrieve values from plan
	var plan authMethodResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// var err error
	plan, diags = r.upsert(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *authMethodResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Updating Auth Method")
	// Retrieve values from plan
	var plan authMethodResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// var err error
	plan, diags = r.upsert(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *authMethodResource) upsert(ctx context.Context, plan authMethodResourceModel) (authMethodResourceModel, diag.Diagnostics) {
	authMethodName := plan.Name.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_auth_method", authMethodName)
	authMethodConfig := waypointClient.DefaultAuthMethodConfig()

	authMethodConfig.Name = authMethodName

	authMethodConfig.Name = plan.Name.ValueString()
	authMethodConfig.DisplayName = plan.DisplayName.ValueString()
	authMethodConfig.Description = plan.Description.ValueString()
	authMethodConfig.AccessSelector = plan.AccessorSelector.ValueString()

	// all auth methods are OIDC at the time
	oidcConfig := waypointClient.DefaultOidcConfig()

	oidcConfig.ClientId = plan.ClientID.ValueString()
	oidcConfig.DiscoveryUrl = plan.DiscoveryURL.ValueString()
	oidcConfig.ClientSecret = plan.ClientSecret.ValueString()

	uris := []string{}
	diags := plan.AllowedRedirectURIs.ElementsAs(ctx, &uris, false)
	if diags.HasError() {
		return plan, diags
	}
	oidcConfig.AllowedRedirectUris = uris

	mappings := make(map[string]string)
	claimDiags := plan.ClaimMappings.ElementsAs(ctx, &mappings, false)
	diags.Append(claimDiags...)
	if diags.HasError() {
		return plan, diags
	}
	oidcConfig.ClaimMappings = mappings

	listMappings := make(map[string]string)
	listDiags := plan.ListClaimMappings.ElementsAs(ctx, &listMappings, false)
	diags.Append(listDiags...)
	if diags.HasError() {
		return plan, diags
	}
	oidcConfig.ListClaimMappings = listMappings

	auds := []string{}
	audsDiags := plan.Auds.ElementsAs(ctx, &auds, false)
	diags.Append(audsDiags...)
	if diags.HasError() {
		return plan, diags
	}
	oidcConfig.Auds = auds

	scopes := []string{}
	scopesDiags := plan.Scopes.ElementsAs(ctx, &scopes, false)
	diags.Append(scopesDiags...)
	if diags.HasError() {
		return plan, diags
	}
	oidcConfig.Scopes = scopes

	signingAlgs := []string{}
	algoDiags := plan.SigningAlgs.ElementsAs(ctx, &signingAlgs, false)
	diags.Append(algoDiags...)
	if diags.HasError() {
		return plan, diags
	}
	oidcConfig.SigningAlgs = signingAlgs

	discoveryCAPEM := []string{}
	discDiags := plan.DiscoveryCAPEM.ElementsAs(ctx, &discoveryCAPEM, false)
	diags.Append(discDiags...)
	if diags.HasError() {
		return plan, diags
	}
	oidcConfig.DiscoveryCaPem = discoveryCAPEM

	_, err := r.client.UpsertOidc(ctx, oidcConfig, authMethodConfig)
	if err != nil {
		diags.AddError(
			"Error updating auth method",
			"Could not update auth method, unexpected error: "+err.Error(),
		)
	}

	return plan, diags
}

// Read refreshes the Terraform state with the latest data.
func (r *authMethodResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Read refreshes the Terraform state with the latest data.
	var state authMethodResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note that the current client we use directly refers to OIDC Auth Methods
	// because at time of writing OIDC was the only auth method available
	getAuthResponse, err := r.client.GetOidcAuthMethod(ctx, state.Name.ValueString())
	if err != nil {
		if status.Code(err) == codes.NotFound {
			tflog.Info(ctx, "Auth Method not found, removing from state")
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Auth Method",
			"Could not read Auth Method with ID"+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	auth := getAuthResponse.AuthMethod
	state.Name = types.StringValue(auth.GetName())
	method := auth.GetOidc()
	state.ClientID = types.StringValue(method.GetClientId())
	state.DiscoveryURL = types.StringValue(method.GetDiscoveryUrl())

	allowedRedirectURIs, diags := types.ListValueFrom(ctx, types.StringType, method.GetAllowedRedirectUris())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.AllowedRedirectURIs = allowedRedirectURIs

	claims, diags := types.MapValueFrom(ctx, types.StringType, method.GetClaimMappings())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ClaimMappings = claims

	listClaims, diags := types.MapValueFrom(ctx, types.StringType, method.GetListClaimMappings())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ListClaimMappings = listClaims

	discoveryCAPEM, diags := types.ListValueFrom(ctx, types.StringType, method.GetDiscoveryCaPem())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DiscoveryCAPEM = discoveryCAPEM

	signingAlgs, diags := types.ListValueFrom(ctx, types.StringType, method.GetSigningAlgs())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.SigningAlgs = signingAlgs

	scopes, diags := types.ListValueFrom(ctx, types.StringType, method.GetScopes())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Scopes = scopes

	auds, diags := types.ListValueFrom(ctx, types.StringType, method.GetAuds())
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Auds = auds

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *authMethodResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Retrieve values from state
	var state authMethodResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	authMethodName := state.Name.ValueString()
	ctx = tflog.SetField(ctx, "waypoint_auth_method", authMethodName)

	// Delete existing auth method
	err := r.client.DeleteOidc(ctx, authMethodName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Waypoint Auth Method",
			"Could not delete auth method, unexpected error: "+err.Error(),
		)
		return
	}
}
