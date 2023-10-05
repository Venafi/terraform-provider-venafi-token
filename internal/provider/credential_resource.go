package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-providers/terraform-provider-venafi-token/internal/model"
	"github.com/terraform-providers/terraform-provider-venafi-token/internal/vcertclient"
)

const (
	// attributes of the resource
	fURL            = "url"
	fUsername       = "username"
	fPassword       = "password"
	fP12Cert        = "p12_cert_filename"
	fP12Password    = "p12_cert_password"
	fAccessToken    = "access_token"
	fRefreshToken   = "refresh_token"
	fClientID       = "client_id"
	fExpirationDate = "expiration"
	fTrustBundle    = "trust_bundle"
	fRefreshWindow  = "refresh_window"

	// messages
	msgCredentialResourceError = "credential resource error"
	msgImportFail              = "failed to import certificate resource"

	// default values
	defaultClientID      = "hashicorp-terraform-by-venafi"
	defaultRefreshWindow = 30 // in days

	resourceNameSuffix = "credential"
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
	resp.TypeName = fmt.Sprintf("%s_%s", req.ProviderTypeName, resourceNameSuffix)
}

func (r *CredentialResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Venafi Credential Resource",

		Attributes: map[string]schema.Attribute{
			fURL: schema.StringAttribute{
				MarkdownDescription: "The Venafi TLSPDC URL. Example: https://tpp.venafi.example/vedsdk",
				Optional:            true,
				Computed:            true,
			},
			fUsername: schema.StringAttribute{
				MarkdownDescription: "Username to authenticate to TLSPDC and request a new token",
				Optional:            true,
				Computed:            true,
			},
			fPassword: schema.StringAttribute{
				MarkdownDescription: "Password to authenticate to TLSPDC and request a new token",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			fP12Cert: schema.StringAttribute{
				MarkdownDescription: "base64-encoded PKCS#12 keystore containing a vcert certificate, private key, and chain certificates to authenticate to TLSPDC",
				Optional:            true,
				Computed:            true,
			},
			fP12Password: schema.StringAttribute{
				MarkdownDescription: "Password for the PKCS#12 keystore declared in p12_cert",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			fAccessToken: schema.StringAttribute{
				MarkdownDescription: "Access token used for authorization to TLSPDC",
				Computed:            true,
				Sensitive:           true,
			},
			fRefreshToken: schema.StringAttribute{
				MarkdownDescription: "Token used to request a new token pair (access/refresh token) from a TLSPDC instance",
				Optional:            true,
				Computed:            true,
				Sensitive:           true,
			},
			fClientID: schema.StringAttribute{
				MarkdownDescription: "Application that will be using the token",
				Optional:            true,
				Computed:            true,
			},
			fExpirationDate: schema.Int64Attribute{
				MarkdownDescription: "Expiration date of the access token",
				Optional:            true,
				Computed:            true,
			},
			fTrustBundle: schema.StringAttribute{
				MarkdownDescription: "Use to specify a base64-encoded, PEM-formatted file that contains certificates to be trust anchors for all communications with the Venafi TLSPDC instance",
				Optional:            true,
				Computed:            true,
			},
			fRefreshWindow: schema.Int64Attribute{
				MarkdownDescription: "number of days before expiration where a token refresh should be done",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

func (r *CredentialResource) Create(_ context.Context, _ resource.CreateRequest, resp *resource.CreateResponse) {
	resp.Diagnostics.AddError(msgCredentialResourceError, "credential resource cannot be created, only imported.")
	return
}

func (r *CredentialResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	tflog.Info(ctx, fmt.Sprintf("reading credential resource"))
	var data model.CredentialResourceData
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// No access token, request a new pair right away
	if data.AccessToken.IsNull() {
		tflog.Info(ctx, "no access token, retrieving a new token pair")
		err := rotateToken(ctx, &data)
		if err != nil {
			reportClientError(ctx, err, resp)
			return
		}
		resp.State.Set(ctx, data)
		return
	}

	// Got access token, check expiration
	client := vcertclient.New(ctx, data)
	expired, err := client.VerifyTokenExpired()
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("client error: %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to verify token expiration, got error: %s", err))
		return
	}

	// If token already expired, request new pair
	if expired {
		tflog.Info(ctx, "access token expired, retrieving a new token pair")
		err = rotateToken(ctx, &data)
		if err != nil {
			reportClientError(ctx, err, resp)
			return
		}
		resp.State.Set(ctx, data)
		return
	}

	// Refresh window is in days, we need to convert it to seconds: n days * 24 hours * 60 minutes * 60 seconds
	refreshWindowSeconds := data.RefreshWindow.ValueInt64() * 24 * 60 * 60
	// If token not expired, check expiration date is on refresh window. If so, request new pair
	if data.ExpirationDate.ValueInt64()-refreshWindowSeconds < time.Now().Unix() {
		tflog.Info(ctx, "access token expiration within refresh window, retrieving a new token pair")
		err = rotateToken(ctx, &data)
		if err != nil {
			reportClientError(ctx, err, resp)
			return
		}

		resp.State.Set(ctx, data)
		return
	}

	// Token is valid, nothing to do here
	tflog.Info(ctx, "access token valid")
	return
}

func (r *CredentialResource) Update(ctx context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	tflog.Info(ctx, "updating credential resource")
}

func (r *CredentialResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	tflog.Info(ctx, "deleting credential resource")
	var state model.CredentialResourceData

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	client := vcertclient.New(ctx, state)
	err := client.RevokeToken()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete credential resource: %s", err.Error()))
		return
	}

	resp.State.RemoveResource(ctx)
	tflog.Info(ctx, "successfully revoked access token")
}

func (r *CredentialResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Info(ctx, "importing credential resource")
	id := req.ID

	dataMap, err := getValuesMap(ctx, id)
	if err != nil {
		details := fmt.Sprintf("%s: %s", msgImportFail, err.Error())
		resp.Diagnostics.AddError(msgCredentialResourceError, details)
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("field map: %v", dataMap))

	data := model.CredentialResourceData{}

	msg := "saving attribute to terraform state: [%s]=%s"
	if val, ok := dataMap[fURL]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fURL, val))
		data.URL = types.StringValue(val)
	}
	if val, ok := dataMap[fUsername]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fUsername, val))
		data.Username = types.StringValue(val)
	}
	if val, ok := dataMap[fPassword]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fPassword, val))
		data.Password = types.StringValue(val)
	}
	if val, ok := dataMap[fP12Cert]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fP12Cert, val))
		data.P12Certificate = types.StringValue(val)
	}
	if val, ok := dataMap[fP12Password]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fP12Password, val))
		data.P12Password = types.StringValue(val)
	}
	if val, ok := dataMap[fAccessToken]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fAccessToken, val))
		data.AccessToken = types.StringValue(val)
	}
	if val, ok := dataMap[fRefreshToken]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fRefreshToken, val))
		data.RefreshToken = types.StringValue(val)
	}
	if val, ok := dataMap[fTrustBundle]; ok {
		tflog.Info(ctx, fmt.Sprintf(msg, fTrustBundle, val))
		data.TrustBundle = types.StringValue(val)
	}

	clientID := defaultClientID
	if val, ok := dataMap[fClientID]; ok {
		clientID = val
	}
	tflog.Info(ctx, fmt.Sprintf(msg, fClientID, clientID))
	data.ClientID = types.StringValue(clientID)

	refreshWindow := defaultRefreshWindow
	if val, ok := dataMap[fRefreshWindow]; ok {
		valInt, err := strconv.Atoi(val)
		if err != nil {
			details := fmt.Sprintf("%s: %s", msgImportFail, err.Error())
			resp.Diagnostics.AddError(msgCredentialResourceError, details)
			return
		}
		refreshWindow = valInt
	}
	tflog.Info(ctx, fmt.Sprintf(msg, fRefreshWindow, fmt.Sprintf("%d", refreshWindow)))
	data.RefreshWindow = types.Int64Value(int64(refreshWindow))

	tflog.Debug(ctx, fmt.Sprintf("data struct: %v", data))
	diags := resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
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

func rotateToken(ctx context.Context, data *model.CredentialResourceData) error {
	client := vcertclient.New(ctx, *data)
	clientResp, err := client.RequestNewTokenPair()
	if err != nil {
		return err
	}

	data.AccessToken = types.StringValue(clientResp.AccessToken)
	data.ExpirationDate = types.Int64Value(clientResp.Expires)
	data.RefreshToken = types.StringValue(clientResp.RefreshToken)

	return nil
}

func reportClientError(ctx context.Context, err error, resp *resource.ReadResponse) {
	tflog.Error(ctx, fmt.Sprintf("client error: %s", err.Error()))
	resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to rotate token, got error: %s", err.Error()))
}
