package oneprovider

import (
    "context"
    "fmt"
    "net/http"
    "net/url"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "encoding/json"
    "strings"
    "io"
)
 
type vmResource struct {
    client *http.Client
    config oneProviderModel
}

type vmResourceModel struct {
    ID           types.String `tfsdk:"id"`
    LocationID   types.String `tfsdk:"location_id"`
    InstanceSize types.String `tfsdk:"instance_size"`
    Template     types.String `tfsdk:"template"`
    Hostname     types.String `tfsdk:"hostname"`
    IPAddress    types.String `tfsdk:"ip_address"`
    Password     types.String `tfsdk:"password"`
    SSHKeys      types.String `tfsdk:"ssh_keys"`
}

func NewVMResource() resource.Resource {
    return &vmResource{
        client: &http.Client{},
    }
}

func (r *vmResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *vmResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
            },
            "location_id": schema.StringAttribute{
                Required: true,
            },
            "instance_size": schema.StringAttribute{
                Required: true,
            },
            "template": schema.StringAttribute{
                Required: true,
            },
            "hostname": schema.StringAttribute{
                Required: true,
            },
            "ip_address": schema.StringAttribute{
                Computed: true,
            },
            "password": schema.StringAttribute{
                Computed:  true,
                Sensitive: true,
            },
            "ssh_keys": schema.StringAttribute{
                Required: true,
            },
        },
    }
}

func (r *vmResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    r.config = req.ProviderData.(oneProviderModel)
}

func (r *vmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan vmResourceModel
    diags := req.Plan.Get(ctx, &plan)
    resp.Diagnostics.Append(diags...)
    if diags.HasError() {
        return
    }

    data := url.Values{
        "location_id":   {plan.LocationID.ValueString()},
        "instance_size": {plan.InstanceSize.ValueString()},
        "template":      {plan.Template.ValueString()},
        "hostname":      {plan.Hostname.ValueString()},
        "ssh_keys":      {plan.SSHKeys.ValueString()},
    }
    httpReq, err := http.NewRequest("POST", "https://api.oneprovider.com/vm/create/", strings.NewReader(data.Encode()))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
        return
    }

    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    httpReq.Header.Set("Api-Key", r.config.APIKey.ValueString())
    httpReq.Header.Set("Client-Key", r.config.ClientKey.ValueString())

    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create VM: %s", err))
        return
    }
    defer httpResp.Body.Close()

    var apiResp struct {
        Result   string `json:"result"`
        Response struct {
            Message    string `json:"message"`
            ID         string `json:"id"`
            IPAddress  string `json:"ip_address"`
            Hostname   string `json:"hostname"`
            Password   string `json:"password"`
        } `json:"response"`
    }
    
    if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to decode response: %s", err))
        return
    }

    if apiResp.Result != "success" {
        resp.Diagnostics.AddError("API Error", "Failed to create VM")
        return
    }

    plan.ID = types.StringValue(apiResp.Response.ID)
    plan.IPAddress = types.StringValue(apiResp.Response.IPAddress)
    plan.Password = types.StringValue(apiResp.Response.Password)

    diags = resp.State.Set(ctx, plan)
    resp.Diagnostics.Append(diags...)
}

func (r *vmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state vmResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if diags.HasError() {
        return
    }

    // Since the API doesn't provide a GET endpoint, we'll just maintain the current state
    // In a real implementation, you'd want to add an API call to refresh the VM status
    diags = resp.State.Set(ctx, &state)
    resp.Diagnostics.Append(diags...)
}

func (r *vmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // Implement if you want to support updating VM properties
    resp.Diagnostics.AddError(
        "Not Implemented",
        "VM updates are not currently supported by this provider",
    )
}

func (r *vmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state vmResourceModel
    diags := req.State.Get(ctx, &state)
    resp.Diagnostics.Append(diags...)
    if diags.HasError() {
        return
    }

    data := url.Values{
        "vm_id": {state.ID.ValueString()},
    }

    httpReq, err := http.NewRequest("POST", "https://api.oneprovider.com/vm/destroy/", strings.NewReader(data.Encode()))
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create request: %s", err))
        return
    }

    httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    httpReq.Header.Set("Api-Key", r.config.APIKey.ValueString())
    httpReq.Header.Set("Client-Key", r.config.ClientKey.ValueString())

    httpResp, err := r.client.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete VM: %s", err))
        return
    }
    defer httpResp.Body.Close()

    // Read the response body for detailed error reporting
    bodyBytes, err := io.ReadAll(httpResp.Body)
    if err != nil {
        resp.Diagnostics.AddError("API Error", fmt.Sprintf("Unable to read response body: %s", err))
        return
    }

    var apiResp struct {
        Result   string `json:"result"`
        Response struct {
            Message              string `json:"message"`
            UsageHours          string `json:"usageHours,omitempty"`
            BandwidthOverusage  string `json:"bandwidthOverusage,omitempty"`
            BandwidthOverusageCost string `json:"bandwidthOverusageCost,omitempty"`
            AdditionalHoursForCharge string `json:"additionalHoursForCharge,omitempty"`
        } `json:"response"`
    }
    
    if err := json.Unmarshal(bodyBytes, &apiResp); err != nil {
        resp.Diagnostics.AddError(
            "API Error",
            fmt.Sprintf("Unable to decode response: %s\nRaw response: %s", err, string(bodyBytes)),
        )
        return
    }

    if apiResp.Result != "success" {
        if strings.Contains(apiResp.Response.Message, "confirm_close") {
            resp.Diagnostics.AddError(
                "Bandwidth Overage Warning",
                fmt.Sprintf("Cannot delete VM due to bandwidth overage. Details:\n"+
                    "Message: %s\n"+
                    "Usage Hours: %s\n"+
                    "Bandwidth Overusage: %s GB\n"+
                    "Cost: %s\n"+
                    "Hours to wait: %s\n"+
                    "To force deletion, add 'force_destroy = true' to the resource",
                    apiResp.Response.Message,
                    apiResp.Response.UsageHours,
                    apiResp.Response.BandwidthOverusage,
                    apiResp.Response.BandwidthOverusageCost,
                    apiResp.Response.AdditionalHoursForCharge,
                ),
            )
        } else {
            resp.Diagnostics.AddError(
                "API Error",
                fmt.Sprintf("Failed to delete VM: %s\nRaw response: %s", apiResp.Response.Message, string(bodyBytes)),
            )
        }
        return
    }
}
