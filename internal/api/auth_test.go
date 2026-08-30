package api

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pathParameterPattern = regexp.MustCompile(`\{[^}]+\}`)

func TestAuthenticationMatchesOpenAPIDocument(t *testing.T) {
	t.Parallel()
	document := loadOpenAPIDocument(t)
	handler := newTestHandler(t)

	type endpoint struct {
		operationID  string
		method       string
		path         string
		requiresAuth bool
	}
	var endpoints []endpoint
	for path, pathItem := range document.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			endpoints = append(endpoints, endpoint{
				operationID:  operation.OperationID,
				method:       strings.ToUpper(method),
				path:         pathParameterPattern.ReplaceAllString(path, "00000000-0000-0000-0000-000000000001"),
				requiresAuth: operationRequiresAuthentication(document, operation),
			})
		}
	}
	sort.Slice(endpoints, func(i, j int) bool { return endpoints[i].operationID < endpoints[j].operationID })

	for _, endpoint := range endpoints {
		t.Run(endpoint.operationID, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequestWithContext(t.Context(), endpoint.method, endpoint.path, nil)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if endpoint.requiresAuth {
				assertUnauthorizedResponse(t, response)
			} else {
				assert.NotEqual(t, http.StatusUnauthorized, response.Code, "public endpoint %s %s", endpoint.method, endpoint.path)
			}
		})
	}
}

func operationRequiresAuthentication(document *openapi3.T, operation *openapi3.Operation) bool {
	security := document.Security
	if operation.Security != nil {
		security = *operation.Security
	}
	return len(security) > 0
}

func TestAuthenticationRejectsInvalidCredentials(t *testing.T) {
	t.Parallel()
	handler := newTestHandler(t)

	for _, tt := range []struct {
		name     string
		password string
	}{
		{name: "missing credentials"},
		{name: "wrong password", password: "wrong-password"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			request := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/v1/notes", nil)
			if tt.password != "" {
				request.SetBasicAuth("admin", tt.password)
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			assertUnauthorizedResponse(t, response)
		})
	}
}

func assertUnauthorizedResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	responseBody := response.Body.String()
	require.Equal(t, http.StatusUnauthorized, response.Code, "body = %q", responseBody)
	assert.Equal(t, `Basic realm="hyperboard"`, response.Header().Get("WWW-Authenticate"))
	assert.Equal(t, "text/plain; charset=utf-8", response.Header().Get("Content-Type"))
	assert.Equal(t, "Unauthorized\n", responseBody)
}
