package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func doRequest(cfg *Config, method, reqURL, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequestWithContext(context.TODO(), method, reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.SetBasicAuth("admin", cfg.AdminPassword)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	return resp, nil
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, body)
	}
	return nil
}

func decodeJSON[T any](resp *http.Response) (T, error) {
	var v T
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode response: %w", err)
	}
	return v, nil
}

type listResponse[T any] struct {
	Items  *[]T    `json:"items"`
	Cursor *string `json:"cursor,omitempty"`
}

func fetchAll[T any](cfg *Config, baseURL string, params url.Values) ([]T, error) {
	var all []T
	for {
		u := baseURL
		if len(params) > 0 {
			u += "?" + params.Encode()
		}

		resp, err := doRequest(cfg, http.MethodGet, u, "", nil)
		if err != nil {
			return nil, err
		}
		defer func() { _ = resp.Body.Close() }()

		if err := checkStatus(resp); err != nil {
			return nil, err
		}

		page, err := decodeJSON[listResponse[T]](resp)
		if err != nil {
			return nil, err
		}

		if page.Items != nil {
			all = append(all, *page.Items...)
		}

		if page.Cursor == nil || *page.Cursor == "" {
			break
		}
		params.Set("cursor", *page.Cursor)
	}
	return all, nil
}
