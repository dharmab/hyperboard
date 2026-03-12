package cli

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dharmab/hyperboard/internal/middleware/auth"
	"github.com/dharmab/hyperboard/pkg/client"
)

func (a *App) NewClient() (*client.ClientWithResponses, error) {
	return client.NewClientWithResponses(
		a.Config.APIURL,
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.SetBasicAuth(auth.AdminUsername, a.Config.AdminPassword)
			return nil
		}),
	)
}

func CheckResponse(statusCode int, body []byte) error {
	if statusCode >= 400 {
		return fmt.Errorf("server returned %d: %s", statusCode, body)
	}
	return nil
}
