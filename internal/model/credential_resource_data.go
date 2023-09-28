// Package model contains all data models used across the provider
package model

import "github.com/hashicorp/terraform-plugin-framework/types"

// CredentialResourceData represents a credential resource
type CredentialResourceData struct {
	URL            types.String `tfsdk:"url"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	P12Certificate types.String `tfsdk:"p12_cert_filename"`
	P12Password    types.String `tfsdk:"p12_cert_password"`
	AccessToken    types.String `tfsdk:"access_token"`
	RefreshToken   types.String `tfsdk:"refresh_token"`
	ClientID       types.String `tfsdk:"client_id"`
	ExpirationDate types.Int64  `tfsdk:"expiration"`
	TrustBundle    types.String `tfsdk:"trust_bundle"`
	RefreshWindow  types.Int64  `tfsdk:"refresh_window"`
}
