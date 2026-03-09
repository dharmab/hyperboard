package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type APIClient struct {
	baseURL  string
	password string
	http     *http.Client
}

func newAPIClient(baseURL, password string) *APIClient {
	return &APIClient{
		baseURL:  baseURL,
		password: password,
		http:     &http.Client{},
	}
}

func (c *APIClient) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth("admin", c.password)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.http.Do(req)
}

func (c *APIClient) getRaw(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodGet, path, nil)
}

func (c *APIClient) head(ctx context.Context, path string) (*http.Response, error) {
	return c.do(ctx, http.MethodHead, path, nil)
}

func (c *APIClient) get(ctx context.Context, path string, out any) error {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Message != "" {
			return fmt.Errorf("API %d: %s", resp.StatusCode, apiErr.Message)
		}
		return fmt.Errorf("API %d: %s", resp.StatusCode, string(body))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *APIClient) getWithQuery(ctx context.Context, path string, query url.Values, out any) error {
	if len(query) > 0 {
		path = path + "?" + query.Encode()
	}
	return c.get(ctx, path, out)
}

func (c *APIClient) post(ctx context.Context, path string, body any, out any) (int, error) {
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if out != nil {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode, nil
}

func (c *APIClient) put(ctx context.Context, path string, body any, out any) (int, error) {
	resp, err := c.do(ctx, http.MethodPut, path, body)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		var apiErr struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(respBody, &apiErr) == nil && apiErr.Message != "" {
			return resp.StatusCode, fmt.Errorf("API %d: %s", resp.StatusCode, apiErr.Message)
		}
		return resp.StatusCode, fmt.Errorf("API %d: %s", resp.StatusCode, string(respBody))
	}
	if out != nil {
		return resp.StatusCode, json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode, nil
}

func (c *APIClient) delete(ctx context.Context, path string) (int, error) {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return 0, err
	}
	_ = resp.Body.Close()
	return resp.StatusCode, nil
}

func (c *APIClient) uploadFile(ctx context.Context, data []byte, contentType string, force bool, out any) (int, []byte, error) {
	uploadURL := c.baseURL + "/api/v1/upload"
	if force {
		uploadURL += "?force=true"
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(data))
	if err != nil {
		return 0, nil, fmt.Errorf("create upload request: %w", err)
	}
	req.SetBasicAuth("admin", c.password)
	req.Header.Set("Content-Type", contentType)
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return resp.StatusCode, body, nil
	}
	if out != nil {
		return resp.StatusCode, body, json.Unmarshal(body, out)
	}
	return resp.StatusCode, body, nil
}
