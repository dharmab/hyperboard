package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dharmab/hyperboard/internal/authmw"
	"github.com/dharmab/hyperboard/internal/client"
)

func newClient(cfg *Config) (*client.ClientWithResponses, error) {
	return client.NewClientWithResponses(
		cfg.APIURL,
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.SetBasicAuth(authmw.AdminUsername, cfg.AdminPassword)
			return nil
		}),
	)
}

func checkResponse(statusCode int, body []byte) error {
	if statusCode >= 400 {
		return fmt.Errorf("server returned %d: %s", statusCode, body)
	}
	return nil
}
