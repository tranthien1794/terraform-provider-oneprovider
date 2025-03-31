package oneprovider

import (
    "context"
    "net/http"
    "github.com/hashicorp/terraform-plugin-framework/provider"
    "github.com/hashicorp/terraform-plugin-framework/provider/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/resource" // Added missing import
)

type oneProvider struct {
    client *http.Client
}

type oneProviderModel struct {
    APIKey    types.String `tfsdk:"api_key"`
    ClientKey types.String `tfsdk:"client_key"`
}

func New() provider.Provider {
    return &oneProvider{
        client: &http.Client{},
    }
}

func (p *oneProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
    resp.TypeName = "oneprovider"
    resp.Version = "0.1.0"
}

func (p *oneProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "api_key": schema.StringAttribute{
                Required:    true,
                Description: "API Key for OneProvider",
                Sensitive:   true,
            },
            "client_key": schema.StringAttribute{
                Required:    true,
                Description: "Client Key for OneProvider",
                Sensitive:   true,
            },
        },
    }
}

func (p *oneProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var config oneProviderModel
    diags := req.Config.Get(ctx, &config)
    resp.Diagnostics.Append(diags...)
    if diags.HasError() {
        return
    }
    
    if config.APIKey.IsNull() || config.ClientKey.IsNull() {
        resp.Diagnostics.AddError(
            "Missing Configuration",
            "API Key and Client Key are required",
        )
        return
    }
    
    resp.DataSourceData = config
    resp.ResourceData = config
}

func (p *oneProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    // Return an empty slice since we don't have any data sources yet
    return []func() datasource.DataSource{}
}

func (p *oneProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewVMResource,
    }
}
