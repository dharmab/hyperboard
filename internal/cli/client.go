package cli

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dharmab/hyperboard/internal/middleware/auth"
	"github.com/dharmab/hyperboard/pkg/client"
)

const apiTimeout = 24 * time.Hour

// NewClient creates an authenticated API client using the app's configuration.
func (a *App) NewClient() (*client.ClientWithResponses, error) {
	return client.NewClientWithResponses(
		a.Config.APIURL,
		client.WithHTTPClient(&http.Client{Timeout: apiTimeout}),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.SetBasicAuth(auth.AdminUsername, a.Config.AdminPassword)
			return nil
		}),
	)
}

// CheckResponseStatus validates that an API response has exactly the expected status.
func CheckResponseStatus(statusCode int, body []byte, expectedStatus int) error {
	if statusCode == expectedStatus {
		return nil
	}
	if len(body) > 0 {
		return fmt.Errorf("protocol error: expected HTTP %d, got HTTP %d: %s", expectedStatus, statusCode, body)
	}
	return fmt.Errorf("protocol error: expected HTTP %d, got HTTP %d", expectedStatus, statusCode)
}

// CheckJSONResponse validates an API response's status and required decoded JSON body.
func CheckJSONResponse[T any](statusCode int, body []byte, expectedStatus int, payload *T) error {
	if err := CheckResponseStatus(statusCode, body, expectedStatus); err != nil {
		return err
	}
	if payload == nil {
		return fmt.Errorf("protocol error: HTTP %d response did not contain the expected JSON body", expectedStatus)
	}
	return nil
}
