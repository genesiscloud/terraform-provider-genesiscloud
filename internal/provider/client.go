package provider

import (
	"context"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/genesiscloud/genesiscloud-go"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type ClientLogger struct {
	retryablehttp.Logger

	ctx context.Context
}

func (l ClientLogger) Printf(format string, a ...interface{}) {
	tflog.Trace(l.ctx, fmt.Sprintf(format, a...))
}

type Client struct {
	*genesiscloud.ClientWithResponses

	PollingInterval time.Duration
}

func (c *Client) PollingWait(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(c.PollingInterval):
		return nil
	}
}

type ClientConfig struct {
	genesiscloud.ClientConfig
	PollingInterval time.Duration
}

func NewClient(ctx context.Context, config ClientConfig) (*Client, error) {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 4
	retryClient.Logger = ClientLogger{ctx: ctx}
	retryClient.CheckRetry = RetryPolicy

	opts := []genesiscloud.ClientOption{
		genesiscloud.WithHTTPClient(retryClient.StandardClient()),
	}

	client, err := genesiscloud.NewGenesisCloudClient(config.ClientConfig, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		ClientWithResponses: client,
		PollingInterval:     config.PollingInterval,
	}, nil
}

type DataSourceWithClient struct {
	client *Client
}

func (d *DataSourceWithClient) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

type ResourceWithClient struct {
	client *Client
}

func (r *ResourceWithClient) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// Adapted from https://github.com/hashicorp/go-retryablehttp/blob/7f638666931a119b8d7a1239e44da11db0d446a4/client.go
var (
	// A regular expression to match the error returned by net/http when the
	// configured number of redirects is exhausted. This error isn't typed
	// specifically so we resort to matching on the error string.
	redirectsErrorRe = regexp.MustCompile(`stopped after \d+ redirects\z`)

	// A regular expression to match the error returned by net/http when the
	// scheme specified in the URL is invalid. This error isn't typed
	// specifically so we resort to matching on the error string.
	schemeErrorRe = regexp.MustCompile(`unsupported protocol scheme`)

	// A regular expression to match the error returned by net/http when the
	// TLS certificate is not trusted. This error isn't typed
	// specifically so we resort to matching on the error string.
	notTrustedErrorRe = regexp.MustCompile(`certificate is not trusted`)
)

func RetryPolicy(ctx context.Context, resp *http.Response, err error) (bool, error) {
	// do not retry on context.Canceled or context.DeadlineExceeded
	if ctx.Err() != nil {
		return false, ctx.Err()
	}

	if err != nil {
		if v, ok := err.(*url.Error); ok {
			// Don't retry if the error was due to too many redirects.
			if redirectsErrorRe.MatchString(v.Error()) {
				return false, v
			}

			// Don't retry if the error was due to an invalid protocol scheme.
			if schemeErrorRe.MatchString(v.Error()) {
				return false, v
			}

			// Don't retry if the error was due to TLS cert verification failure.
			if notTrustedErrorRe.MatchString(v.Error()) {
				return false, v
			}
			if _, ok := v.Err.(x509.UnknownAuthorityError); ok {
				return false, v
			}
		}

		// The error is likely recoverable so retry.
		return true, nil
	}

	// 429 Too Many Requests is recoverable. Sometimes the server puts
	// a Retry-After response header to indicate when the server is
	// available to start processing request from client.
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, nil
	}

	// Check the response code. This will also catch
	// invalid response codes as well, like 0 and 999.
	if resp.StatusCode == 0 || resp.StatusCode >= 600 {
		return true, fmt.Errorf("unexpected HTTP status %s", resp.Status)
	}

	return false, nil
}
