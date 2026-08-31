package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"uuid"

	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPutTagCategoryValidation(t *testing.T) {
	t.Parallel()
	server := newTestServer(t)

	for _, name := range []string{"-bad", "_bad", " bad", "!bad", "bad ", "bad  category"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := types.TagCategory{Name: name, Description: "test", Color: "#ff0000"}
			encodedBody, err := json.Marshal(body)
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tagCategories/"+url.PathEscape(name), bytes.NewReader(encodedBody))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			server.PutTagCategory(response, request, name)
			assert.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestPutTagCategoryColorValidation(t *testing.T) {
	t.Parallel()
	server := newTestServer(t)

	for _, color := range []string{"banana", "ff0000", "#fff", "#gggggg"} {
		t.Run(color, func(t *testing.T) {
			t.Parallel()
			name := "color-test-" + uuid.NewV4().String()[:8]
			body := types.TagCategory{Name: name, Description: "test", Color: color}
			encodedBody, err := json.Marshal(body)
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tagCategories/"+name, bytes.NewReader(encodedBody))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			server.PutTagCategory(response, request, name)
			assert.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestCreateTagCategory(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	categoryName := testCategoryName()

	response := performRequest(handler, http.MethodPut, categoryPath(categoryName), categoryBody(categoryName, "created", "#123456"), true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusCreated, response.Code, "body: %s", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var category types.TagCategory
	decodeJSON(t, response.Body.Bytes(), &category)
	assertTagCategoryRequiredFields(t, category, categoryName, "created", "#123456")
}

func TestGetTagCategory(t *testing.T) {
	t.Parallel()
	handler, category := createTagCategoryFixture(t)

	response := performRequest(handler, http.MethodGet, categoryPath(category.Name), "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var fetched types.TagCategory
	decodeJSON(t, response.Body.Bytes(), &fetched)
	assert.Equal(t, category, fetched)
}

func TestListTagCategories(t *testing.T) {
	t.Parallel()
	handler, category := createTagCategoryFixture(t)

	response := performRequest(handler, http.MethodGet, "/api/v1/tagCategories", "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var categories TagCategoriesResponse
	decodeJSON(t, response.Body.Bytes(), &categories)
	require.NotNil(t, categories.Items)
	require.Len(t, *categories.Items, 1)
	listedCategory := (*categories.Items)[0]
	assertTagCategoryRequiredFields(t, listedCategory, category.Name, category.Description, category.Color)
	require.NotNil(t, listedCategory.TagCount)
	assert.Zero(t, *listedCategory.TagCount)
}

func TestUpdateTagCategory(t *testing.T) {
	t.Parallel()
	handler, category := createTagCategoryFixture(t)

	response := performRequest(handler, http.MethodPut, categoryPath(category.Name), categoryBody(category.Name, "updated", "#654321"), true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var updated types.TagCategory
	decodeJSON(t, response.Body.Bytes(), &updated)
	assertTagCategoryRequiredFields(t, updated, category.Name, "updated", "#654321")
	assert.Equal(t, category.CreatedAt, updated.CreatedAt)
	assert.False(t, updated.UpdatedAt.Before(category.UpdatedAt))
}

func TestDeleteTagCategory(t *testing.T) {
	t.Parallel()
	handler, category := createTagCategoryFixture(t)

	response := performRequest(handler, http.MethodDelete, categoryPath(category.Name), "", true)
	assertEmptyNoContent(t, response.Code, response.Body.String())

	missingResponse := performRequest(handler, http.MethodGet, categoryPath(category.Name), "", true)
	assert.Equal(t, http.StatusNotFound, missingResponse.Code, "body: %s", missingResponse.Body.String())

	deleteMissingResponse := performRequest(handler, http.MethodDelete, categoryPath(category.Name), "", true)
	assert.Equal(t, http.StatusNotFound, deleteMissingResponse.Code, "body: %s", deleteMissingResponse.Body.String())
}

func createTagCategoryFixture(t *testing.T) (http.Handler, types.TagCategory) {
	t.Helper()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	categoryName := testCategoryName()
	response := performRequest(handler, http.MethodPut, categoryPath(categoryName), categoryBody(categoryName, "created", "#123456"), true)
	require.Equal(t, http.StatusCreated, response.Code, "body: %s", response.Body.String())
	var category types.TagCategory
	decodeJSON(t, response.Body.Bytes(), &category)
	return handler, category
}

func testCategoryName() string {
	return "category-" + uuid.NewV4().String()[:8]
}

func categoryPath(name string) string {
	return "/api/v1/tagCategories/" + name
}

func categoryBody(name, description, color string) string {
	return `{"name":"` + name + `","description":"` + description + `","color":"` + color + `"}`
}

func assertTagCategoryRequiredFields(t *testing.T, category types.TagCategory, name, description, color string) {
	t.Helper()
	assert.Equal(t, name, category.Name)
	assert.Equal(t, description, category.Description)
	assert.Equal(t, color, category.Color)
	assert.False(t, category.CreatedAt.IsZero())
	assert.False(t, category.UpdatedAt.IsZero())
}

func TestTagCategoriesPagination(t *testing.T) {
	t.Parallel()

	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	want := []string{"category-00", "category-01", "category-02", "category-03", "category-04"}
	for _, name := range want {
		response := performRequest(handler, http.MethodPut, "/api/v1/tagCategories/"+name, `{"name":"`+name+`","color":"#123456"}`, true)
		responseBody := response.Body.String()
		require.Equalf(t, http.StatusCreated, response.Code, "create category %q status = %d; body = %q", name, response.Code, responseBody)
	}

	categoryNames := traverseNamedPages(t, handler, "/api/v1/tagCategories", "", func(body []byte) ([]string, *string) {
		var response TagCategoriesResponse
		decodeJSON(t, body, &response)
		names := make([]string, len(*response.Items))
		for i, item := range *response.Items {
			names[i] = item.Name
		}
		return names, response.Cursor
	})
	assert.Equalf(t, want, categoryNames, "traversed categories = %v, want %v", categoryNames, want)

	assertStaleNamedCursor(t, handler, "/api/v1/tagCategories", "category-015", want[2], func(body []byte) string {
		var response TagCategoriesResponse
		decodeJSON(t, body, &response)
		return (*response.Items)[0].Name
	})
}
