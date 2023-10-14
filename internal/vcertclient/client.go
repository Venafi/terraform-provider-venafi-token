// Package vcertclient contains all functions that interface with vcert-sdk
package vcertclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"

	"github.com/Venafi/vcert/v5"
	"github.com/Venafi/vcert/v5/pkg/endpoint"
	"github.com/Venafi/vcert/v5/pkg/venafi/tpp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/crypto/pkcs12"

	"github.com/terraform-providers/terraform-provider-venafi-token/internal/model"
)

const (
	msgTokenRefreshStart   = "retrieving token pair using"
	msgTokenRefreshSuccess = "successfully retrieved new token pair"
	msgTokenRefreshFail    = "failed to retrieve new token pair with"
	msgVcertClientError    = "terraform vcert client error"
)

type Client struct {
	context  context.Context
	credData model.CredentialResourceData
}

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
	Expires      int64
	ExpiresIn    int64
}

func New(ctx context.Context, data model.CredentialResourceData) *Client {
	return &Client{
		context:  ctx,
		credData: data,
	}
}

func (c *Client) VerifyTokenExpired() (expired bool, err error) {
	tflog.Info(c.context, "verifying access token validity")

	config, err := c.createVCertConfig()
	if err != nil {
		tflog.Error(c.context, err.Error())
		return false, err
	}

	vClient, err := vcert.NewClient(config, false)
	if err != nil {
		tflog.Error(c.context, err.Error())
		return false, err
	}

	auth := &endpoint.Authentication{
		AccessToken: c.credData.AccessToken.ValueString(),
	}

	//Due to limitations in TPP API, we cannot retrieve the access token expiration time from the verify function
	_, err = vClient.(*tpp.Connector).VerifyAccessToken(auth)
	if err != nil {
		msg := fmt.Sprintf("%s: %s", msgVcertClientError, err.Error())
		tflog.Info(c.context, msg)
		return true, nil
	}

	return false, nil
}

func (c *Client) RequestNewTokenPair() (*RefreshTokenResponse, error) {
	tflog.Info(c.context, "requesting new token pair")

	tokenMethod := !c.credData.RefreshToken.IsNull()
	p12Method := !c.credData.P12Certificate.IsNull() && !c.credData.P12Password.IsNull()
	userMethod := !c.credData.Username.IsNull() && !c.credData.Password.IsNull()

	if !tokenMethod && !p12Method && !userMethod {
		return nil, fmt.Errorf("%s: no authorization methods specified", msgVcertClientError)
	}

	if tokenMethod {
		tflog.Info(c.context, fmt.Sprintf("%s %s", msgTokenRefreshStart, "refresh token"))
		resp, err := c.refreshAccessToken()
		// return if no errors
		if err == nil {
			tflog.Info(c.context, msgTokenRefreshSuccess)
			return resp, nil
		}
		// if refresh token fails. Check if there is any other auth method.
		// if there is another auth method, log warning and continue
		msg := fmt.Sprintf("%s %s: %s", msgTokenRefreshFail, "refresh token", err.Error())
		if !p12Method && !userMethod {
			// no other auth method. Log and return error
			tflog.Error(c.context, msg)
			return nil, fmt.Errorf("%s: %w", msgVcertClientError, err)
		}
		// print warning and let other auth methods be used
		tflog.Warn(c.context, msg)
	}

	if p12Method {
		tflog.Info(c.context, fmt.Sprintf("%s %s", msgTokenRefreshStart, "client certificate"))
		resp, err := c.getAccessTokenByP12()
		// return if no errors
		if err == nil {
			tflog.Info(c.context, msgTokenRefreshSuccess)
			return resp, nil
		}
		// if client certificate fails. Check if there is user/password method, log warning and continue
		msg := fmt.Sprintf("%s %s: %s", msgTokenRefreshFail, "client certificate", err.Error())
		if !userMethod {
			// no other auth method. Log and return error
			tflog.Error(c.context, msg)
			return nil, fmt.Errorf("%s: %w", msgVcertClientError, err)
		}
		// log warning and let other auth methods be used
		tflog.Warn(c.context, msg)
	}

	if userMethod {
		tflog.Info(c.context, fmt.Sprintf("%s %s", msgTokenRefreshStart, "username-password"))
		resp, err := c.getAccessTokenByUsernamePassword()
		// return if no errors
		if err == nil {
			tflog.Info(c.context, msgTokenRefreshSuccess)
			return resp, nil
		}
		// no other auth method. Log and return error
		tflog.Error(c.context, fmt.Sprintf("%s %s: %s", msgTokenRefreshFail, "username-password", err.Error()))
		return nil, fmt.Errorf("%s: %w", msgVcertClientError, err)
	}

	return nil, fmt.Errorf("%s: could not complete refresh token operation: all authentication methods failed", msgVcertClientError)
}

func (c *Client) RevokeToken() error {
	tflog.Info(c.context, "revoking access token")

	config, err := c.createVCertConfig()
	if err != nil {
		tflog.Error(c.context, err.Error())
		return err
	}

	vClient, err := vcert.NewClient(config, false)
	if err != nil {
		tflog.Error(c.context, err.Error())
		return err
	}

	auth := &endpoint.Authentication{
		AccessToken: c.credData.AccessToken.ValueString(),
	}
	err = vClient.(*tpp.Connector).RevokeAccessToken(auth)
	if err != nil {
		tflog.Error(c.context, err.Error())
		return err
	}

	return nil
}

func (c *Client) refreshAccessToken() (*RefreshTokenResponse, error) {
	tflog.Info(c.context, "using refresh token authentication method")

	config, err := c.createVCertConfig()
	if err != nil {
		return nil, err
	}
	vClient, err := vcert.NewClient(config, false)
	if err != nil {
		return nil, err
	}

	auth := &endpoint.Authentication{
		RefreshToken: c.credData.RefreshToken.ValueString(),
		ClientId:     c.credData.ClientID.ValueString(),
	}
	resp, err := vClient.(*tpp.Connector).RefreshAccessToken(auth)
	if err != nil {
		return nil, err
	}

	refreshResp := RefreshTokenResponse{
		AccessToken:  resp.Access_token,
		RefreshToken: resp.Refresh_token,
		Expires:      int64(resp.Expires),
	}

	return &refreshResp, nil
}

func (c *Client) getAccessTokenByP12() (*RefreshTokenResponse, error) {
	tflog.Info(c.context, "using client certificate authentication method")

	err := c.configureTLSClient()
	if err != nil {
		return nil, err
	}

	return c.getAccessToken(true)
}

func (c *Client) getAccessTokenByUsernamePassword() (*RefreshTokenResponse, error) {
	tflog.Info(c.context, "using username-password authentication method")

	return c.getAccessToken(false)
}

func (c *Client) getAccessToken(useClientCertificate bool) (*RefreshTokenResponse, error) {
	config, err := c.createVCertConfig()
	if err != nil {
		return nil, err
	}

	vClient, err := vcert.NewClient(config, false)
	if err != nil {
		return nil, err
	}

	auth := &endpoint.Authentication{
		ClientId: c.credData.ClientID.ValueString(),
	}

	if useClientCertificate {
		auth.ClientPKCS12 = true
	} else {
		auth.User = c.credData.Username.ValueString()
		auth.Password = c.credData.Password.ValueString()
	}

	resp, err := vClient.(*tpp.Connector).GetRefreshToken(auth)
	if err != nil {
		return nil, err
	}

	refreshResp := RefreshTokenResponse{
		AccessToken:  resp.Access_token,
		RefreshToken: resp.Refresh_token,
		Expires:      int64(resp.Expires),
	}
	return &refreshResp, nil
}

func (c *Client) configureTLSClient() error {
	tflog.Info(c.context, "configuring TLS client")

	p12Location := c.credData.P12Certificate.ValueString()
	password := c.credData.P12Password.ValueString()

	data, err := os.ReadFile(p12Location)
	if err != nil {
		return fmt.Errorf("%s: unable to read PKCS#12 file at [%s]: %w", msgVcertClientError, p12Location, err)
	}

	// We have a PKCS12 file to use, set it up for cert authentication
	blocks, err := pkcs12.ToPEM(data, password)
	if err != nil {
		return fmt.Errorf("%s: failed converting PKCS#12 archive file to PEM blocks: %w", msgVcertClientError, err)
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	// Construct TLS certificate from PEM data
	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		return fmt.Errorf("%s: failed reading PEM data to build X.509 certificate: %w", msgVcertClientError, err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(pemData)

	// Setup TLS configuration
	tlsConfig := tls.Config{
		Renegotiation: tls.RenegotiateFreelyAsClient,
		Certificates:  []tls.Certificate{cert},
		RootCAs:       caCertPool,
	}

	// Create own Transport to allow HTTP1.1 connections
	transport := &http.Transport{
		// Only one request is made with a client
		DisableKeepAlives: true,
		// This is to allow for http1.1 connections
		ForceAttemptHTTP2: false,
		TLSClientConfig:   &tlsConfig,
	}

	//Setting Default HTTP Transport
	http.DefaultTransport = transport

	tflog.Info(c.context, "TLS client configured")
	return nil
}

func (c *Client) createVCertConfig() (*vcert.Config, error) {
	config := vcert.Config{
		ConnectorType: endpoint.ConnectorTypeTPP,
		BaseUrl:       c.credData.URL.ValueString(),
		LogVerbose:    true,
	}

	if !c.credData.TrustBundle.IsNull() {
		location := c.credData.TrustBundle.ValueString()
		data, err := os.ReadFile(location)
		if err != nil {
			return nil, fmt.Errorf("%s: unable to read trust bundle file at [%s]: %w", msgVcertClientError, location, err)
		}
		config.ConnectionTrust = string(data)
	}

	return &config, nil
}
