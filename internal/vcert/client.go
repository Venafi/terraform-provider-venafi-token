package vcert

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net/http"

	"github.com/Venafi/vcert/v5"
	"github.com/Venafi/vcert/v5/pkg/endpoint"
	"github.com/Venafi/vcert/v5/pkg/venafi/tpp"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/crypto/pkcs12"
)

type Client struct {
	context         context.Context
	config          vcert.Config
	trustBundle     string
	useClientPKCS12 bool
}

func New(ctx context.Context, config vcert.Config, trustBundle string) *Client {
	return &Client{
		context:         ctx,
		config:          config,
		trustBundle:     trustBundle,
		useClientPKCS12: false,
	}
}

func (c *Client) ConfigureTLSClient(certificate string, p12Password string) error {
	tflog.Info(c.context, fmt.Sprintf("configuring TLS client"))

	tlsConfig := tls.Config{}

	tlsConfig.Renegotiation = tls.RenegotiateFreelyAsClient

	// We have a PKCS12 file to use, set it up for cert authentication
	blocks, err := pkcs12.ToPEM([]byte(certificate), p12Password)
	if err != nil {
		return fmt.Errorf("failed converting PKCS#12 archive file to PEM blocks: %w", err)
	}

	var pemData []byte
	for _, b := range blocks {
		pemData = append(pemData, pem.EncodeToMemory(b)...)
	}

	// Construct TLS certificate from PEM data
	cert, err := tls.X509KeyPair(pemData, pemData)
	if err != nil {
		return fmt.Errorf("failed reading PEM data to build X.509 certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(pemData)

	// Setup HTTPS client
	tlsConfig.Certificates = []tls.Certificate{cert}
	tlsConfig.RootCAs = caCertPool

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

	c.useClientPKCS12 = true
	tflog.Info(c.context, "TLS client configured")
	return nil
}

type RefreshResponse struct {
	AccessToken  string
	RefreshToken string
	Expires      int
	ExpiresIn    int64
}

func (c *Client) RequestNewToken() (*RefreshResponse, error) {
	tflog.Info(c.context, "requesting new token pair")

	tokenMethod := false
	if c.config.Credentials.RefreshToken != "" {
		tokenMethod = true
	}

	p12Method := c.useClientPKCS12

	userMethod := false
	if c.config.Credentials.User != "" && c.config.Credentials.Password != "" {
		userMethod = true
	}

	if !tokenMethod && !p12Method && !userMethod {
		return nil, fmt.Errorf("no authorization methods specified - cannot get a new access token")
	}

	var err error

	if tokenMethod {
		resp, err := c.refreshAccessToken()
		// Return if no errors. Otherwise, let the other authorization methods be used
		if err == nil {
			tflog.Info(c.context, "successfully retrieved new refresh token")
			return resp, nil
		}
	}

	if p12Method {
		resp, err := c.getAccessTokenByP12()
		// Return if no errors. Otherwise, let the other authorization methods be used
		if err == nil {
			tflog.Info(c.context, "successfully retrieved new refresh token")
			return resp, nil
		}
	}

	if userMethod {
		resp, err := c.getAccessTokenByUsernamePassword()
		// Return if no errors. Otherwise, let the other authorization methods be used
		if err == nil {
			tflog.Info(c.context, "successfully retrieved new refresh token")
			return resp, nil
		}
	}

	return nil, fmt.Errorf("could not complete refresh token operation: %w", err)
}

func (c *Client) refreshAccessToken() (*RefreshResponse, error) {
	vClient, err := vcert.NewClient(&c.config, false)
	if err != nil {
		return nil, err
	}

	tflog.Info(c.context, "using refresh token")

	auth := &endpoint.Authentication{
		RefreshToken: c.config.Credentials.RefreshToken,
		ClientId:     c.config.Credentials.ClientId,
	}
	resp, err := vClient.(*tpp.Connector).RefreshAccessToken(auth)
	if err != nil {
		tflog.Error(c.context, "failed to refresh TLSPDC tokens", map[string]interface{}{"error": err})
		return nil, err
	}

	return &RefreshResponse{
		AccessToken:  resp.Access_token,
		RefreshToken: resp.Refresh_token,
		Expires:      resp.Expires,
	}, nil

}

func (c *Client) getAccessTokenByP12() (*RefreshResponse, error) {
	return c.getAccessToken(true)
}

func (c *Client) getAccessTokenByUsernamePassword() (*RefreshResponse, error) {
	return c.getAccessToken(false)
}

func (c *Client) getAccessToken(useP12Cert bool) (*RefreshResponse, error) {
	vClient, err := vcert.NewClient(&c.config, false)
	if err != nil {
		return nil, err
	}

	auth := &endpoint.Authentication{
		ClientId: c.config.Credentials.ClientId,
		Scope:    c.config.Credentials.Scope,
	}
	if useP12Cert {
		auth.ClientPKCS12 = true
	} else {
		auth.User = c.config.Credentials.User
		auth.Password = c.config.Credentials.Password
	}

	resp, err := vClient.(*tpp.Connector).GetRefreshToken(auth)
	if err != nil {
		tflog.Error(c.context, "failed to refresh TLSPDC tokens", map[string]interface{}{"error": err})
		return nil, err
	}

	return &RefreshResponse{
		AccessToken:  resp.Access_token,
		RefreshToken: resp.Refresh_token,
		Expires:      resp.Expires,
	}, nil
}
