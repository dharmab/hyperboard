package main

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/internal/middleware/auth"
)

func newAPIClient(baseURL, password string) (*client.ClientWithResponses, error) {
	return client.NewClientWithResponses(
		baseURL,
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.SetBasicAuth(auth.AdminUsername, password)
			return nil
		}),
	)
}

// mediaClient handles raw HTTP requests for media proxying,
// which is not part of the OpenAPI spec.
type mediaClient struct {
	baseURL  string
	password string
	http     *http.Client
}

func newMediaClient(baseURL, password string) *mediaClient {
	return &mediaClient{
		baseURL:  baseURL,
		password: password,
		http:     &http.Client{},
	}
}

func (c *mediaClient) getRaw(ctx context.Context, path string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth("admin", c.password)
	return c.http.Do(req)
}

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
	if resp.StatusCode != http.StatusOK {
		http.Error(w, "Media not found", resp.StatusCode)
		return
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if cl := resp.Header.Get("Content-Length"); cl != "" {
		w.Header().Set("Content-Length", cl)
	}
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = io.Copy(w, resp.Body)
}
