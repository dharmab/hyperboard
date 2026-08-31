package web

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/dharmab/hyperboard/internal/middleware/auth"
	"github.com/dharmab/hyperboard/pkg/client"
)

const (
	// outboundOperationTimeout bounds complete API and media transfers while leaving
	// enough time for multi-gigabyte uploads and downloads on slow connections.
	outboundOperationTimeout = 24 * time.Hour
	outboundDialTimeout      = 30 * time.Second
	outboundTLSHandshake     = 30 * time.Second
	outboundResponseHeaders  = 5 * time.Minute
)

func newOutboundTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   outboundDialTimeout,
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.TLSHandshakeTimeout = outboundTLSHandshake
	transport.ResponseHeaderTimeout = outboundResponseHeaders
	return transport
}

func newOutboundHTTPClient() *http.Client {
	return &http.Client{
		Transport: newOutboundTransport(),
		Timeout:   outboundOperationTimeout,
	}
}

// newAPIClient creates an authenticated OpenAPI client for the Hyperboard API.
func newAPIClient(baseURL, password string) (*client.ClientWithResponses, error) {
	return client.NewClientWithResponses(
		baseURL,
		client.WithHTTPClient(newOutboundHTTPClient()),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.SetBasicAuth(auth.AdminUsername, password)
			return nil
		}),
	)
}

// mediaClient handles streaming media requests for the web proxy.
type mediaClient struct {
	baseURL  string
	password string
	http     *http.Client
}

// newMediaClient creates an HTTP client for proxying media requests to the API.
func newMediaClient(baseURL, password string) *mediaClient {
	return &mediaClient{
		baseURL:  baseURL,
		password: password,
		http:     newOutboundHTTPClient(),
	}
}

// getRaw performs an authenticated GET request to the API.
func (c *mediaClient) getRaw(ctx context.Context, path, byteRange string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth("admin", c.password)
	if byteRange != "" {
		req.Header.Set("Range", byteRange)
	}
	return c.http.Do(req)
}

// head performs an authenticated HEAD request to the API.
func (c *mediaClient) head(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth("admin", c.password)
	return c.http.Do(req)
}

// copyMediaResponse streams a media response to the HTTP writer.
func copyMediaResponse(w http.ResponseWriter, resp *http.Response) {
	for _, name := range []string{
		"Content-Type",
		"Content-Length",
		"Content-Disposition",
		"Cache-Control",
		"X-Content-Type-Options",
		"Accept-Ranges",
		"Content-Range",
	} {
		if value := resp.Header.Get(name); value != "" {
			w.Header().Set(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
