package api

import (
	"errors"
	"net/http"
	"slices"
	"testing"
	"time"
	"uuid"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagRelationshipIntegrityAndCounts(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	suffix := uuid.NewV4().String()[:8]
	categoryName := "relationship-category-" + suffix
	unrelatedCategoryName := "relationship-unrelated-category-" + suffix
	deletedTagName := "relationship-deleted-tag-" + suffix
	siblingTagName := "relationship-sibling-tag-" + suffix
	unrelatedTagName := "relationship-unrelated-tag-" + suffix

	for _, name := range []string{categoryName, unrelatedCategoryName} {
		response := performRequest(handler, http.MethodPut, "/api/v1/tagCategories/"+name, `{"name":"`+name+`","color":"#123456"}`, true)
		responseBody := response.Body.String()
		require.Equal(t, http.StatusCreated, response.Code, "body: %s", responseBody)
	}
	for _, tt := range []struct {
		name     string
		category string
	}{
		{name: deletedTagName, category: categoryName},
		{name: siblingTagName, category: categoryName},
		{name: unrelatedTagName, category: unrelatedCategoryName},
	} {
		body := `{"name":"` + tt.name + `","category":"` + tt.category + `"}`
		response := performRequest(handler, http.MethodPut, "/api/v1/tags/"+tt.name, body, true)
		createTagResponseBody := response.Body.String()
		require.Equal(t, http.StatusCreated, response.Code, "body: %s", createTagResponseBody)
	}

	postWithDeletedTag := createTestPost(t, sqlStore)
	siblingOnlyPost := createTestPost(t, sqlStore)
	_, err := sqlStore.UpdatePost(t.Context(), uuid.UUID(postWithDeletedTag.ID), "", []string{deletedTagName, siblingTagName}, time.Now().UTC())
	require.NoError(t, err)
	_, err = sqlStore.UpdatePost(t.Context(), uuid.UUID(siblingOnlyPost.ID), "", []string{siblingTagName}, time.Now().UTC())
	require.NoError(t, err)

	assertRoutedTagCounts(t, handler, map[string]int{deletedTagName: 1, siblingTagName: 2, unrelatedTagName: 0})
	assertRoutedCategoryCounts(t, handler, map[string]int{categoryName: 2, unrelatedCategoryName: 1})

	deleteTagResponse := performRequest(handler, http.MethodDelete, "/api/v1/tags/"+deletedTagName, "", true)
	assertEmptyNoContent(t, deleteTagResponse.Code, deleteTagResponse.Body.String())
	response := performRequest(handler, http.MethodGet, "/api/v1/tags/"+deletedTagName, "", true)
	deletedTagResponseBody := response.Body.String()
	assert.Equal(t, http.StatusNotFound, response.Code, "body: %s", deletedTagResponseBody)
	assertPostTagNames(t, sqlStore, uuid.UUID(postWithDeletedTag.ID), []string{siblingTagName})
	assertPostTagNames(t, sqlStore, uuid.UUID(siblingOnlyPost.ID), []string{siblingTagName})
	assertRoutedTagCounts(t, handler, map[string]int{siblingTagName: 2, unrelatedTagName: 0})
	assertRoutedCategoryCounts(t, handler, map[string]int{categoryName: 1, unrelatedCategoryName: 1})

	deleteCategoryResponse := performRequest(handler, http.MethodDelete, "/api/v1/tagCategories/"+categoryName, "", true)
	assertEmptyNoContent(t, deleteCategoryResponse.Code, deleteCategoryResponse.Body.String())
	response = performRequest(handler, http.MethodGet, "/api/v1/tagCategories/"+categoryName, "", true)
	deletedCategoryResponseBody := response.Body.String()
	assert.Equal(t, http.StatusNotFound, response.Code, "body: %s", deletedCategoryResponseBody)
	siblingResponse := performRequest(handler, http.MethodGet, "/api/v1/tags/"+siblingTagName, "", true)
	siblingResponseBody := siblingResponse.Body.String()
	require.Equal(t, http.StatusOK, siblingResponse.Code, "body: %s", siblingResponseBody)
	var sibling types.Tag
	decodeJSON(t, siblingResponse.Body.Bytes(), &sibling)
	assert.Nil(t, sibling.Category)
	unrelatedResponse := performRequest(handler, http.MethodGet, "/api/v1/tags/"+unrelatedTagName, "", true)
	unrelatedResponseBody := unrelatedResponse.Body.String()
	require.Equal(t, http.StatusOK, unrelatedResponse.Code, "body: %s", unrelatedResponseBody)
	var unrelated types.Tag
	decodeJSON(t, unrelatedResponse.Body.Bytes(), &unrelated)
	require.NotNil(t, unrelated.Category)
	assert.Equal(t, unrelatedCategoryName, *unrelated.Category)
	assertRoutedCategoryCounts(t, handler, map[string]int{unrelatedCategoryName: 1})
	assertPostTagNames(t, sqlStore, uuid.UUID(postWithDeletedTag.ID), []string{siblingTagName})
}

func assertRoutedTagCounts(t *testing.T, handler http.Handler, want map[string]int) {
	t.Helper()
	response := performRequest(handler, http.MethodGet, "/api/v1/tags", "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var tags TagsResponse
	decodeJSON(t, response.Body.Bytes(), &tags)
	require.NotNil(t, tags.Items)
	require.Len(t, *tags.Items, len(want))
	for _, tag := range *tags.Items {
		count := want[tag.Name]
		if !assert.Contains(t, want, tag.Name, "listed tag") {
			continue
		}
		require.NotNil(t, tag.PostCount)
		assert.Equal(t, count, *tag.PostCount)
	}
}

func assertRoutedCategoryCounts(t *testing.T, handler http.Handler, want map[string]int) {
	t.Helper()
	response := performRequest(handler, http.MethodGet, "/api/v1/tagCategories", "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var categories TagCategoriesResponse
	decodeJSON(t, response.Body.Bytes(), &categories)
	require.NotNil(t, categories.Items)
	require.Len(t, *categories.Items, len(want))
	for _, category := range *categories.Items {
		count := want[category.Name]
		if !assert.Contains(t, want, category.Name, "listed category") {
			continue
		}
		require.NotNil(t, category.TagCount)
		assert.Equal(t, count, *category.TagCount)
	}
}

func assertPostTagNames(t *testing.T, sqlStore store.SQLStore, postID uuid.UUID, want []string) {
	t.Helper()
	post, err := sqlStore.GetPost(t.Context(), postID)
	require.NoError(t, err)
	postTagNames := make([]string, len(post.Tags))
	for i, tag := range post.Tags {
		postTagNames[i] = tag.Name
	}
	assert.True(t, slices.Equal(postTagNames, want))
}

func TestConvertTagToAlias(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	suffix := uuid.NewV4().String()[:8]
	sourceName := "source-" + suffix
	targetName := "target-" + suffix
	sourceAlias := "source-alias-" + suffix
	targetAlias := "target-alias-" + suffix

	for _, body := range []string{
		`{"name":"` + sourceName + `","aliases":["` + sourceAlias + `"]}`,
		`{"name":"` + targetName + `","aliases":["` + targetAlias + `"]}`,
	} {
		var tagInput types.Tag
		decodeJSON(t, []byte(body), &tagInput)
		response := performRequest(handler, http.MethodPut, "/api/v1/tags/"+tagInput.Name, body, true)
		createTagResponseBody := response.Body.String()
		require.Equal(t, http.StatusCreated, response.Code, "body: %s", createTagResponseBody)
	}

	sourceOnlyPost := createTestPost(t, sqlStore)
	bothTagsPost := createTestPost(t, sqlStore)
	_, err := sqlStore.UpdatePost(t.Context(), uuid.UUID(sourceOnlyPost.ID), "", []string{sourceName}, time.Now().UTC())
	require.NoError(t, err)
	_, err = sqlStore.UpdatePost(t.Context(), uuid.UUID(bothTagsPost.ID), "", []string{sourceName, targetName}, time.Now().UTC())
	require.NoError(t, err)

	response := performRequest(handler, http.MethodPost, "/api/v1/tags/"+sourceName+"/convert-to-alias", `{"target":"`+targetName+`"}`, true)
	conversionResponseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", conversionResponseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var converted types.Tag
	decodeJSON(t, response.Body.Bytes(), &converted)
	require.Equal(t, targetName, converted.Name)
	require.NotNil(t, converted.Aliases)
	for _, alias := range []string{sourceName, sourceAlias, targetAlias} {
		assert.Contains(t, *converted.Aliases, alias)
	}

	for _, postID := range []uuid.UUID{uuid.UUID(sourceOnlyPost.ID), uuid.UUID(bothTagsPost.ID)} {
		post, err := sqlStore.GetPost(t.Context(), postID)
		require.NoError(t, err)
		require.Len(t, post.Tags, 1)
		assert.Equal(t, targetName, post.Tags[0].Name)
	}

	listResponse := performRequest(handler, http.MethodGet, "/api/v1/tags", "", true)
	listResponseBody := listResponse.Body.String()
	require.Equal(t, http.StatusOK, listResponse.Code, "body: %s", listResponseBody)
	var tags TagsResponse
	decodeJSON(t, listResponse.Body.Bytes(), &tags)
	require.NotNil(t, tags.Items)
	require.Len(t, *tags.Items, 1)
	assert.Equal(t, targetName, (*tags.Items)[0].Name)
}

func TestConvertTagToAliasPreservesStateAfterStoreFailure(t *testing.T) {
	t.Parallel()

	suffix := uuid.NewV4().String()[:8]
	sourceName := "fault-source-" + suffix
	targetName := "fault-target-" + suffix
	sourceAlias := "fault-source-alias-" + suffix
	targetAlias := "fault-target-alias-" + suffix
	baseStore := newTestStore(t)
	setupHandler := newTestHandlerForServer(t, NewServer(baseStore, memory.New()))
	for _, body := range []string{
		`{"name":"` + sourceName + `","aliases":["` + sourceAlias + `"]}`,
		`{"name":"` + targetName + `","aliases":["` + targetAlias + `"]}`,
	} {
		var input types.Tag
		decodeJSON(t, []byte(body), &input)
		response := performRequest(setupHandler, http.MethodPut, "/api/v1/tags/"+input.Name, body, true)
		responseBody := response.Body.String()
		require.Equal(t, http.StatusCreated, response.Code, "create tag %q body: %s", input.Name, responseBody)
	}

	sourceOnlyPost := createTestPost(t, baseStore)
	bothTagsPost := createTestPost(t, baseStore)
	_, err := baseStore.UpdatePost(t.Context(), uuid.UUID(sourceOnlyPost.ID), "", []string{sourceName}, time.Now().UTC())
	require.NoError(t, err, "tag source-only post")
	_, err = baseStore.UpdatePost(t.Context(), uuid.UUID(bothTagsPost.ID), "", []string{sourceName, targetName}, time.Now().UTC())
	require.NoError(t, err, "tag post with source and target")

	faultStore := faultInjectingSQLStore{SQLStore: baseStore, convertTagToAliasErr: errors.New("injected conversion failure")}
	response := performRequest(newTestHandlerForServer(t, NewServer(faultStore, memory.New())), http.MethodPost, "/api/v1/tags/"+sourceName+"/convert-to-alias", `{"target":"`+targetName+`"}`, true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusInternalServerError, response.Code, "convert tag body: %s", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))

	source, err := baseStore.GetTag(t.Context(), sourceName)
	require.NoError(t, err, "get source tag after failed conversion")
	target, err := baseStore.GetTag(t.Context(), targetName)
	require.NoError(t, err, "get target tag after failed conversion")
	aliases, err := baseStore.GetTagAliases(t.Context(), uuid.UUID(source.ID), uuid.UUID(target.ID))
	require.NoError(t, err, "get aliases after failed conversion")
	sourceAliases := aliases[uuid.UUID(source.ID)]
	require.Len(t, sourceAliases, 1, "source aliases")
	assert.Equal(t, sourceAlias, sourceAliases[0], "source alias")
	targetAliases := aliases[uuid.UUID(target.ID)]
	require.Len(t, targetAliases, 1, "target aliases")
	assert.Equal(t, targetAlias, targetAliases[0], "target alias")

	for _, tt := range []struct {
		postID uuid.UUID
		want   []string
	}{
		{postID: uuid.UUID(sourceOnlyPost.ID), want: []string{sourceName}},
		{postID: uuid.UUID(bothTagsPost.ID), want: []string{sourceName, targetName}},
	} {
		post, err := baseStore.GetPost(t.Context(), tt.postID)
		require.NoError(t, err, "get post %s after failed conversion", tt.postID)
		postTagNames := make([]string, len(post.Tags))
		for i, tag := range post.Tags {
			postTagNames[i] = tag.Name
		}
		assert.Equal(t, tt.want, postTagNames, "post %s tags", tt.postID)
	}
}
