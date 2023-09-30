package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var _ provider.Provider = &VenafiTokenProvider{}

func New() provider.Provider {
	return &VenafiTokenProvider{}
}

type VenafiTokenProvider struct{}

func (p *VenafiTokenProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "venafi-token"
}

func (p *VenafiTokenProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This is for refreshing Venafi tokens for use with venafi-provider.",
	}
}

func (p *VenafiTokenProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {

}

func (p *VenafiTokenProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (p *VenafiTokenProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCredentialResource,
	}
}
