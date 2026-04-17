package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/realianx/terraform-provider-resend/resend"
)

var _ provider.Provider = &ResendProvider{}

type ResendProvider struct {
	version string
}

type ResendProviderModel struct {
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ResendProvider{version: version}
	}
}

func (p *ResendProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "resend"
	resp.Version = p.version
}

func (p *ResendProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the Resend email API.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Resend API key. Can also be set via RESEND_API_KEY environment variable.",
			},
		},
	}
}

func (p *ResendProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config ResendProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiKey := os.Getenv("RESEND_API_KEY")
	if !config.APIKey.IsNull() && !config.APIKey.IsUnknown() {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing Resend API Key",
			"Set the api_key attribute or RESEND_API_KEY environment variable.",
		)
		return
	}

	client := resend.NewClient(apiKey)
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *ResendProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		ResendTemplateResource,
		ResendSegmentResource,
		ResendTopicResource,
		ResendAutomationResource,
		ResendEventResource,
	}
}

func (p *ResendProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
