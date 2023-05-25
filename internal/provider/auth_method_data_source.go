// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	waypointClient "github.com/hashicorp-dev-advocates/waypoint-client/pkg/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &authMethodDataSource{}
	_ datasource.DataSourceWithConfigure = &authMethodDataSource{}
)

// NewAuthMethodDataSource is a helper function to simplify the provider implementation.
func NewAuthMethodDataSource() datasource.DataSource {
	return &authMethodDataSource{}
}

// authMethodDataSource is the data source implementation.
type authMethodDataSource struct {
	client waypointClient.Waypoint
}

// Configure adds the provider configured client to the data source.
func (d *authMethodDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(waypointClient.Waypoint)
}

// Metadata returns the data source type name.
func (d *authMethodDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth_method"
}

// authMethodDataModel maps the data schema data.
type authMethodDataModel struct {
	Name                types.String `tfsdk:"name"`
	DisplayName         types.String `tfsdk:"display_name"`
	Description         types.String `tfsdk:"description"`
	AccessorSelector    types.String `tfsdk:"accessor_selector"`
	ClientID            types.String `tfsdk:"client_id"`
	DiscoveryURL        types.String `tfsdk:"discovery_url"`
	AllowedRedirectURIs types.List   `tfsdk:"allowed_redirect_uris"`
	ClaimMappings       types.Map    `tfsdk:"claim_mappings"`
	ListClaimMappings   types.Map    `tfsdk:"list_claim_mappings"`
	DiscoveryCAPEM      types.List   `tfsdk:"discovery_ca_pem"`
	SigningAlgs         types.List   `tfsdk:"signing_algs"`
	Scopes              types.List   `tfsdk:"scopes"`
	Auds                types.List   `tfsdk:"auds"`
}

// Schema defines the schema for the data source.
func (d *authMethodDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// Schema defines the schema for the data source.
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the Auth Method",
			},
			"display_name": schema.StringAttribute{
				Computed:    true,
				Description: "The display name of the Auth Method",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of auth method",
			},
			"accessor_selector": schema.StringAttribute{
				Computed: true,
			},
			"client_id": schema.StringAttribute{
				Computed:    true,
				Description: "Client ID of OIDC provider",
			},
			"discovery_url": schema.StringAttribute{
				Computed:    true,
				Description: "Discovery URL for OIDC provider",
			},
			"allowed_redirect_uris": schema.ListAttribute{
				Computed:    true,
				Description: "Allowed URI for auth redirection.",
				ElementType: types.StringType,
			},
			"claim_mappings": schema.MapAttribute{
				Computed:    true,
				Description: "Mapping of a claim to a variable value for the access selector",
				ElementType: types.StringType,
			},
			"list_claim_mappings": schema.MapAttribute{
				Computed:    true,
				Description: "Same as claim-mapping but for list values",
				ElementType: types.StringType,
			},
			"discovery_ca_pem": schema.ListAttribute{
				Computed:    true,
				Description: "Optional CA certificate chain to validate the discovery URL. Multiple CA certificates can be specified to support easier rotation",
				ElementType: types.StringType,
			},
			"signing_algs": schema.ListAttribute{
				Computed:    true,
				Description: "The signing algorithms supported by the OIDC connect server. If this isn't specified, this will default to RS256 since that should be supported according to the RFC. The string values here should be valid OIDC signing algorithms",
				ElementType: types.StringType,
			},
			"scopes": schema.ListAttribute{
				Computed:    true,
				Description: "The optional claims scope requested.",
				ElementType: types.StringType,
			},
			"auds": schema.ListAttribute{
				Computed:    true,
				Description: "The optional audience claims required",
				ElementType: types.StringType,
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *authMethodDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Read refreshes the Terraform state with the latest data.
	var state authMethodDataModel
	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Note that the current client we use directly refers to OIDC Auth Methods
	// because at time of writing OIDC was the only auth method available
	getAuthResponse, err := d.client.GetOidcAuthMethod(context.TODO(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Auth Method",
			"Could not read Auth Method with ID"+state.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	auth := getAuthResponse.AuthMethod
	state.Name = types.StringValue(auth.GetName())
	state.DisplayName = types.StringValue(auth.GetDisplayName())
	state.Description = types.StringValue(auth.GetDescription())
	state.AccessorSelector = types.StringValue(auth.GetAccessSelector())
	state.AccessorSelector = types.StringValue(auth.GetAccessSelector())
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
