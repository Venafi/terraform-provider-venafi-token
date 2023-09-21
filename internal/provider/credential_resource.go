package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	// Attributes of the resource
	fURL            = "url"
	fUsername       = "username"
	fPassword       = "password"
	fP12Cert        = "p12_cert"
	fP12Password    = "p12_password"
	fAccessToken    = "access_token"
	fRefreshToken   = "refresh_token"
	fClientID       = "client_id"
	fExpirationDate = "expiration"
	fTrustBundle    = "trust_bundle"
)

var (
	_ resource.Resource                = &CredentialResource{}
	_ resource.ResourceWithImportState = &CredentialResource{}
)

func NewCredentialResource() resource.Resource {
	return &CredentialResource{}
}

type CredentialResource struct{}

func (r *CredentialResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = "venafi_credential"
}

func (r *CredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Venafi Credential Resource",

		Attributes: map[string]schema.Attribute{
			fURL: schema.StringAttribute{
				MarkdownDescription: "The Venafi TLSPDC URL. Example: https://tpp.venafi.example/vedsdk",
				//Optional:            true,
				Required: true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fUsername: schema.StringAttribute{
				MarkdownDescription: "Username to authenticate to TLSPDC and request a new token",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fPassword: schema.StringAttribute{
				MarkdownDescription: "Password to authenticate to TLSPDC and request a new token",
				Optional:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fP12Cert: schema.StringAttribute{
				MarkdownDescription: "base64-encoded PKCS#12 keystore containing a vcert certificate, private key, and chain certificates to authenticate to TLSPDC",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fP12Password: schema.StringAttribute{
				MarkdownDescription: "Password for the PKCS#12 keystore declared in p12_cert",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fAccessToken: schema.StringAttribute{
				MarkdownDescription: "Access token used for authorization to TLSPDC",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fRefreshToken: schema.StringAttribute{
				MarkdownDescription: "Token used to request a new token pair (access/refresh token) from a TLSPDC instance",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fClientID: schema.StringAttribute{
				MarkdownDescription: "Application that will be using the token",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
			fExpirationDate: schema.Int64Attribute{
				MarkdownDescription: "Expiration date of the access token",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					timeoutUnknownModifier{},
				},
			},
			fTrustBundle: schema.StringAttribute{
				MarkdownDescription: "Use to specify a base64-encoded, PEM-formatted file that contains certificates to be trust anchors for all communications with the Venafi TLSPDC instance",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					timeoutUnknownModifier{},
				},
			},
		},
	}
}

type credentialResourceData struct {
	URL            types.String `tfsdk:"url"`
	Username       types.String `tksdk:"username"`
	Password       types.String `tfsdk:"password"`
	P12Certificate types.String `tfsdk:"p12_cert"`
	P12Password    types.String `tfsdk:"p12_password"`
	AccessToken    types.String `tfsdk:"acccess_token"`
	RefreshToken   types.String `tfsdk:"refresh_token"`
	ClientID       types.String `tfsdk:"client_id"`
	ExpirationDate types.Int64  `tfsdk:"expiration"`
	TrustBundle    types.String `tfsdk:"trust_bundle"`
}

func (r *CredentialResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError("credential resource error", "credential resource cannot be created only imported.")
	return
}

func (r *CredentialResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
	// Not possible
}

func (r *CredentialResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//var state credentialResourceData
	//var data credentialResourceData
	//
	//diags := req.Plan.Get(ctx, &data)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//diags = req.State.Get(ctx, &state)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//client := client.New()
	//var resultJson refreshResponse
	//err := client.Request(ctx, "tooling.tokens.rotate?refresh_token="+state.RefreshToken.ValueString(), &resultJson)
	//if err != nil {
	//	resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotate token, got error: %s", err))
	//	return
	//}
	//
	//data.Token = types.StringValue(resultJson.Token)
	//data.Expires = types.Int64Value(resultJson.Expires - 60*60*3)
	//data.RefreshToken = types.StringValue(resultJson.RefreshToken)
	//
	//diags = resp.State.Set(ctx, &data)
	//resp.Diagnostics.Append(diags...)
}

func (r *CredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *CredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("refresh_token"), req, resp)
	id := req.ID

	dataMap, err := getValuesMap(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}
	data := &credentialResourceData{}

	msg := "saving %s attribute to terraform state"
	if val, ok := dataMap[fURL]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fURL))
		data.URL = types.StringValue(val)
	}
	if val, ok := dataMap[fUsername]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fUsername))
		data.Username = types.StringValue(val)
	}
	if val, ok := dataMap[fPassword]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fPassword))
		data.Password = types.StringValue(val)
	}
	if val, ok := dataMap[fP12Cert]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fP12Cert))
		data.P12Certificate = types.StringValue(val)
	}
	if val, ok := dataMap[fP12Password]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fP12Password))
		data.P12Password = types.StringValue(val)
	}
	if val, ok := dataMap[fAccessToken]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fAccessToken))
		data.AccessToken = types.StringValue(val)
	}
	if val, ok := dataMap[fRefreshToken]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fRefreshToken))
		data.RefreshToken = types.StringValue(val)
	}
	if val, ok := dataMap[fClientID]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fClientID))
		data.ClientID = types.StringValue(val)
	}
	if val, ok := dataMap[fTrustBundle]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fTrustBundle))
		data.TrustBundle = types.StringValue(val)
	}

	resp.State.Set(ctx, data)
}

type timeoutUnknownModifier struct{}

func (m timeoutUnknownModifier) Description(ctx context.Context) string {
	return "Allow token refresh before token expires."
}

func (m timeoutUnknownModifier) MarkdownDescription(ctx context.Context) string {
	return "Allow token refresh before token expires."
}

func (m timeoutUnknownModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	var data credentialResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if data.ExpirationDate.ValueInt64() < time.Now().Unix() {
		resp.PlanValue = types.Int64Unknown()
	}
}

func (m timeoutUnknownModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	var data credentialResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if data.ExpirationDate.ValueInt64() < time.Now().Unix() {
		resp.PlanValue = types.StringUnknown()

	}
}

func getValuesMap(ctx context.Context, values string) (map[string]string, error) {

	dict := make(map[string]string)

	list := strings.Split(values, ",")
	for _, item := range list {
		key, value, found := strings.Cut(item, "=")
		if !found {
			msg := fmt.Sprintf("no separator found on value: %s", item)
			tflog.Info(ctx, msg)
			return nil, fmt.Errorf(msg)
		}
		tflog.Debug(ctx, fmt.Sprintf("credential field found: %s = %s", key, value))
		dict[key] = value
	}

	return dict, nil
}
