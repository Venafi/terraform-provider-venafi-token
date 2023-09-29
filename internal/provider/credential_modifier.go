package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/terraform-providers/terraform-provider-venafi-token/internal/model"
	"github.com/terraform-providers/terraform-provider-venafi-token/internal/vcertclient"
)

var _ planmodifier.String = &credentialUnknownModifier{}

type credentialUnknownModifier struct{}

func (m credentialUnknownModifier) Description(_ context.Context) string {
	return "allow token refresh before token expires."
}

func (m credentialUnknownModifier) MarkdownDescription(_ context.Context) string {
	return "allow token refresh before token expires."
}

func (m credentialUnknownModifier) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	tflog.Info(ctx, fmt.Sprintf("running venafi_credential plan modifier for %s", req.Path.String()))
	var data model.CredentialResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Mark resource for update if there is no access token
	if data.AccessToken.IsNull() {
		tflog.Info(ctx, "no access token, marking resource for update")
		resp.PlanValue = types.Int64Unknown()
		return
	}

	client := vcertclient.New(ctx, data)
	expired, expirationDate, err := client.VerifyTokenExpired()
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("client error: %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to verify token, got error: %s", err))
		return
	}

	tflog.Info(ctx, fmt.Sprintf("token state: expired=%t, expirationDate=%d", expired, expirationDate))

	// If token already expired, mark resource for update
	if expired {
		tflog.Info(ctx, "access token expired, marking resource for update")
		resp.PlanValue = types.Int64Unknown()
		return
	}

	// If token not expired, check expiration date is on refresh window. If so, mark resource for update
	refreshWindowSeconds := data.ExpirationDate.ValueInt64() * 24 * 60 * 60
	if *expirationDate < (time.Now().Unix() - refreshWindowSeconds) {
		tflog.Info(ctx, "access token expiration within refresh window, marking resource for update")
		resp.PlanValue = types.Int64Unknown()
	}
}

func (m credentialUnknownModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	tflog.Info(ctx, fmt.Sprintf("running venafi_credential plan modifier for %s", req.Path.String()))

	var data model.CredentialResourceData
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Mark resource for update if there is no access token
	if data.AccessToken.IsNull() {
		tflog.Info(ctx, "no access token, marking resource for update")
		resp.PlanValue = types.StringUnknown()
		return
	}

	client := vcertclient.New(ctx, data)
	expired, expirationDate, err := client.VerifyTokenExpired()
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("client error: %s", err.Error()))
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to verify token, got error: %s", err))
		return
	}
	tflog.Info(ctx, fmt.Sprintf("token state: expired=%t, expirationDate=%d", expired, expirationDate))

	// If token already expired, mark resource for update
	if expired {
		tflog.Info(ctx, "access token expired, marking resource for update")
		resp.PlanValue = types.StringUnknown()
		return
	}

	// If token not expired, check expiration date is on refresh window. If so, mark resource for update
	refreshWindowSeconds := data.ExpirationDate.ValueInt64() * 24 * 60 * 60
	if *expirationDate < (time.Now().Unix() - refreshWindowSeconds) {
		tflog.Info(ctx, "access token expiration within refresh window, marking resource for update")
		resp.PlanValue = types.StringUnknown()
	}
}
