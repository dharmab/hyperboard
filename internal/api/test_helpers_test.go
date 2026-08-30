package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/dharmab/hyperboard/internal/middleware/auth"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/legacy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPassword = "test-password"

func init() {
	binaryDecoder := func(body io.Reader, _ http.Header, _ *openapi3.SchemaRef, _ openapi3filter.EncodingFn) (any, error) {
		contents, err := io.ReadAll(body)
		return string(contents), err
	}
	for _, contentType := range []string{"image/png", "image/gif", "image/webp", "video/mp4"} {
		openapi3filter.RegisterBodyDecoder(contentType, binaryDecoder)
	}
}

func newTestHandler(t *testing.T) http.Handler {
	t.Helper()
	return newTestHandlerForServer(t, newTestServer(t))
}

func newTestHandlerForServer(t *testing.T, server *Server) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	HandlerWithOptions(server, StdHTTPServerOptions{
		BaseRouter:       mux,
		ErrorHandlerFunc: parameterBindingErrorHandler,
	})
	handler := auth.BasicAuthMiddleware(
		testPassword,
		"/healthz",
		"/readyz",
		"/metrics",
	)(mux)
	return newOpenAPIValidationHandler(t, handler)
}

func loadOpenAPIDocument(t *testing.T) *openapi3.T {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	require.True(t, ok, "locate test helper")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	document, err := loader.LoadFromFile(filepath.Join(filepath.Dir(filename), "spec", "api-spec.yaml"))
	require.NoError(t, err, "load OpenAPI specification")
	require.NoError(t, document.Validate(t.Context()), "validate OpenAPI specification")
	return document
}

func newOpenAPIValidationHandler(t *testing.T, handler http.Handler) http.Handler {
	t.Helper()
	document := loadOpenAPIDocument(t)
	router, err := legacy.NewRouter(document)
	require.NoError(t, err, "create OpenAPI router")

	return http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		requestBody, err := io.ReadAll(request.Body)
		assert.NoError(t, err, "read %s %s request body for OpenAPI validation", request.Method, request.URL.Path)
		request.Body = io.NopCloser(bytes.NewReader(requestBody))

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)

		route, pathParams, routeErr := router.FindRoute(request)
		if routeErr != nil {
			assert.Contains(t, []int{http.StatusNotFound, http.StatusMethodNotAllowed}, recorder.Code,
				"match %s %s to OpenAPI operation: %v", request.Method, request.URL.Path, routeErr)
		} else {
			request.Body = io.NopCloser(bytes.NewReader(requestBody))
			input := &openapi3filter.RequestValidationInput{
				Request:    request,
				PathParams: pathParams,
				Route:      route,
				Options: &openapi3filter.Options{
					AuthenticationFunc: func(context.Context, *openapi3filter.AuthenticationInput) error { return nil },
				},
			}
			if shouldValidateOpenAPIRequest(recorder.Code) {
				err := openapi3filter.ValidateRequest(request.Context(), input)
				assert.NoError(t, err, "validate %s %s OpenAPI request", request.Method, request.URL.Path)
			}

			responseInput := &openapi3filter.ResponseValidationInput{
				RequestValidationInput: input,
				Status:                 recorder.Code,
				Header:                 recorder.Header(),
				Body:                   io.NopCloser(bytes.NewReader(recorder.Body.Bytes())),
			}
			err := openapi3filter.ValidateResponse(request.Context(), responseInput)
			assert.NoError(t, err, "validate %s %s OpenAPI response", request.Method, request.URL.Path)
		}

		maps.Copy(response.Header(), recorder.Header())
		response.WriteHeader(recorder.Code)
		if recorder.Code != http.StatusNoContent {
			_, err := response.Write(recorder.Body.Bytes())
			assert.NoError(t, err, "copy validated response")
		}
	})
}

func shouldValidateOpenAPIRequest(status int) bool {
	switch status {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusRequestEntityTooLarge,
		http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity:
		return false
	default:
		return true
	}
}

func performRequest(handler http.Handler, method, target, body string, authenticated bool) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(context.Background(), method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if authenticated {
		req.SetBasicAuth("any-user", testPassword)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func decodeJSON(t *testing.T, body []byte, destination any) {
	t.Helper()
	require.NoError(t, json.Unmarshal(body, destination))
}

func assertJSONContentType(t *testing.T, contentType string) {
	t.Helper()
	assert.Equal(t, "application/json", contentType)
}

func assertEmptyNoContent(t *testing.T, status int, body string) {
	t.Helper()
	require.Equal(t, http.StatusNoContent, status)
	assert.Empty(t, body)
}
