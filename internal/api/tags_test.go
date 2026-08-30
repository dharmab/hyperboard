package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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

func TestIsValidName(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		valid bool
	}{
		{"abc", true},
		{"ABC", true},
		{"123", true},
		{"1abc", true},
		{"café", true},
		{"日本語", true}, //nolint:gosmopolitan // intentional: testing Unicode letter support
		{"", false},
		{"-abc", false},
		{"_abc", false},
		{" abc", false},
		{"!abc", false},
		{"abc ", false},
		{"abc  def", false},
		{"abc def", true},
		{"abc\t\tdef", false},
		{"abc\tdef", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			valid := isValidName(tc.name)
			assert.Equal(t, tc.valid, valid)
		})
	}
}

func TestIsValidHexColor(t *testing.T) {
	t.Parallel()
	cases := []struct {
		input string
		valid bool
	}{
		{"#ff0000", true},
		{"#FF0000", true},
		{"#000000", true},
		{"#ffffff", true},
		{"#aAbBcC", true},
		{"", false},
		{"ff0000", false},
		{"#fff", false},
		{"#gggggg", false},
		{"banana", false},
		{"#1234567", false},
		{"#12345", false},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			valid := isValidHexColor(tc.input)
			assert.Equal(t, tc.valid, valid)
		})
	}
}

func TestPutTagValidation(t *testing.T) {
	t.Parallel()
	server := newTestServer(t)

	for _, name := range []string{"-bad", "_bad", " bad", "!bad", "bad ", "bad  tag"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			body := types.Tag{Name: name, Description: "test"}
			encodedBody, err := json.Marshal(body)
			require.NoError(t, err)
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/api/v1/tags/"+url.PathEscape(name), bytes.NewReader(encodedBody))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			server.PutTag(response, request, name)
			assert.Equal(t, http.StatusBadRequest, response.Code)
		})
	}
}

func TestTagRenameCollisionPreservesState(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	suffix := uuid.NewV4().String()[:8]
	source := "rename-source-" + suffix
	target := "rename-target-" + suffix
	putConflictTag(t, handler, source, `{"name":"`+source+`","description":"source"}`, http.StatusCreated)
	putConflictTag(t, handler, target, `{"name":"`+target+`","description":"target"}`, http.StatusCreated)

	response := putConflictTag(t, handler, source, `{"name":"`+target+`","description":"changed"}`, http.StatusConflict)
	assertConflictResponse(t, response)
	assertConflictTag(t, handler, source, "source", nil)
	assertConflictTag(t, handler, target, "target", nil)
}

func TestTagCategoryRenameCollisionPreservesState(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	suffix := uuid.NewV4().String()[:8]
	source := "category-source-" + suffix
	target := "category-target-" + suffix
	putConflictCategory(t, handler, source, `{"name":"`+source+`","description":"source","color":"#112233"}`, http.StatusCreated)
	putConflictCategory(t, handler, target, `{"name":"`+target+`","description":"target","color":"#445566"}`, http.StatusCreated)

	response := putConflictCategory(t, handler, source, `{"name":"`+target+`","description":"changed","color":"#ffffff"}`, http.StatusConflict)
	assertConflictResponse(t, response)
	assertConflictCategory(t, handler, source, "source", "#112233")
	assertConflictCategory(t, handler, target, "target", "#445566")
}

func TestConvertTagToAliasConflictSemanticsPreserveState(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		sourceMode string
		targetMode string
		wantStatus int
	}{
		{name: "source equals target", sourceMode: "exists", targetMode: "same", wantStatus: http.StatusBadRequest},
		{name: "missing source", sourceMode: "missing", targetMode: "exists", wantStatus: http.StatusNotFound},
		{name: "missing target", sourceMode: "exists", targetMode: "missing", wantStatus: http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
			suffix := uuid.NewV4().String()[:8]
			source := "convert-source-" + suffix
			target := "convert-target-" + suffix
			if tt.sourceMode == "exists" {
				putConflictTag(t, handler, source, `{"name":"`+source+`","description":"source","aliases":["source-alias-`+suffix+`"]}`, http.StatusCreated)
			}
			switch tt.targetMode {
			case "exists":
				putConflictTag(t, handler, target, `{"name":"`+target+`","description":"target","aliases":["target-alias-`+suffix+`"]}`, http.StatusCreated)
			case "same":
				target = source
			}

			response := performRequest(handler, http.MethodPost, "/api/v1/tags/"+source+"/convert-to-alias", `{"target":"`+target+`"}`, true)
			responseBody := response.Body.String()
			require.Equal(t, tt.wantStatus, response.Code, "case %q; body: %s", tt.name, responseBody)
			if tt.sourceMode == "exists" {
				assertConflictTag(t, handler, source, "source", []string{"source-alias-" + suffix})
			}
			if tt.targetMode == "exists" {
				assertConflictTag(t, handler, target, "target", []string{"target-alias-" + suffix})
			}
		})
	}
}

func TestTagAliasNameCollisionsPreserveState(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name string
		run  func(*testing.T, http.Handler, string)
	}{
		{name: "alias conflicts with tag name", run: func(t *testing.T, handler http.Handler, suffix string) {
			t.Helper()
			source := "alias-source-" + suffix
			target := "alias-target-" + suffix
			oldAlias := "old-alias-" + suffix
			putConflictTag(t, handler, source, `{"name":"`+source+`","description":"source","aliases":["`+oldAlias+`"]}`, http.StatusCreated)
			putConflictTag(t, handler, target, `{"name":"`+target+`","description":"target"}`, http.StatusCreated)
			response := putConflictTag(t, handler, source, `{"name":"`+source+`","description":"changed","aliases":["`+target+`"]}`, http.StatusConflict)
			assertConflictResponse(t, response)
			assertConflictTag(t, handler, source, "source", []string{oldAlias})
			assertConflictTag(t, handler, target, "target", nil)
		}},
		{name: "tag name conflicts with alias", run: func(t *testing.T, handler http.Handler, suffix string) {
			t.Helper()
			source := "name-source-" + suffix
			owner := "name-owner-" + suffix
			alias := "claimed-alias-" + suffix
			putConflictTag(t, handler, source, `{"name":"`+source+`","description":"source"}`, http.StatusCreated)
			putConflictTag(t, handler, owner, `{"name":"`+owner+`","description":"owner","aliases":["`+alias+`"]}`, http.StatusCreated)
			response := putConflictTag(t, handler, source, `{"name":"`+alias+`","description":"changed"}`, http.StatusConflict)
			assertConflictResponse(t, response)
			assertConflictTag(t, handler, source, "source", nil)
			assertConflictTag(t, handler, owner, "owner", []string{alias})
		}},
		{name: "alias conflicts with another alias", run: func(t *testing.T, handler http.Handler, suffix string) {
			t.Helper()
			source := "duplicate-source-" + suffix
			owner := "duplicate-owner-" + suffix
			oldAlias := "duplicate-old-" + suffix
			claimedAlias := "duplicate-claimed-" + suffix
			putConflictTag(t, handler, source, `{"name":"`+source+`","description":"source","aliases":["`+oldAlias+`"]}`, http.StatusCreated)
			putConflictTag(t, handler, owner, `{"name":"`+owner+`","description":"owner","aliases":["`+claimedAlias+`"]}`, http.StatusCreated)
			response := putConflictTag(t, handler, source, `{"name":"`+source+`","description":"changed","aliases":["`+claimedAlias+`"]}`, http.StatusConflict)
			assertConflictResponse(t, response)
			assertConflictTag(t, handler, source, "source", []string{oldAlias})
			assertConflictTag(t, handler, owner, "owner", []string{claimedAlias})
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.run(t, newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New())), uuid.NewV4().String()[:8])
		})
	}
}

func TestConcurrentTagNameAndAliasClaimsConflict(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	handler := newTestHandlerForServer(t, NewServer(sqlStore, memory.New()))
	suffix := uuid.NewV4().String()[:8]
	owner := "concurrent-owner-" + suffix
	claimedName := "concurrent-claim-" + suffix
	putConflictTag(t, handler, owner, `{"name":"`+owner+`"}`, http.StatusCreated)

	start := make(chan struct{})
	statuses := make(chan int, 2)
	go func() {
		<-start
		response := performRequest(handler, http.MethodPut, "/api/v1/tags/"+claimedName, `{"name":"`+claimedName+`"}`, true)
		statuses <- response.Code
	}()
	go func() {
		<-start
		response := performRequest(handler, http.MethodPut, "/api/v1/tags/"+owner, `{"name":"`+owner+`","aliases":["`+claimedName+`"]}`, true)
		statuses <- response.Code
	}()
	close(start)

	gotStatuses := []int{<-statuses, <-statuses}
	conflictCount := 0
	for _, status := range gotStatuses {
		if status == http.StatusConflict {
			conflictCount++
		}
	}
	assert.Equal(t, 1, conflictCount, "operation statuses: %v", gotStatuses)
	assert.True(t, slices.Contains(gotStatuses, http.StatusOK) || slices.Contains(gotStatuses, http.StatusCreated), "operation statuses: %v", gotStatuses)

	_, tagErr := sqlStore.GetTag(t.Context(), claimedName)
	ownerTag, err := sqlStore.GetTag(t.Context(), owner)
	require.NoError(t, err)
	aliases, err := sqlStore.GetTagAliases(t.Context(), uuid.UUID(ownerTag.ID))
	require.NoError(t, err)
	ownerAliases := aliases[uuid.UUID(ownerTag.ID)]
	assert.NotEqual(t, tagErr == nil, slices.Contains(ownerAliases, claimedName), "name must be canonical or an alias, never both")
}

func TestDuplicateTagAliases(t *testing.T) {
	t.Parallel()

	t.Run("creation is rejected without persisting the tag", func(t *testing.T) {
		t.Parallel()
		handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
		suffix := uuid.NewV4().String()[:8]
		name := "duplicate-create-" + suffix
		alias := "duplicate-alias-" + suffix

		response := putConflictTag(t, handler, name, `{"name":"`+name+`","aliases":["`+alias+`","`+alias+`"]}`, http.StatusConflict)
		assertConflictResponse(t, response)

		getResponse := performRequest(handler, http.MethodGet, "/api/v1/tags/"+name, "", true)
		require.Equal(t, http.StatusNotFound, getResponse.Code, "duplicate-alias tag should not be persisted; body: %s", getResponse.Body.String())
	})

	t.Run("update is rejected without changing existing state", func(t *testing.T) {
		t.Parallel()
		handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
		suffix := uuid.NewV4().String()[:8]
		name := "duplicate-update-" + suffix
		oldAlias := "duplicate-old-" + suffix
		duplicateAlias := "duplicate-new-" + suffix
		putConflictTag(t, handler, name, `{"name":"`+name+`","description":"original","aliases":["`+oldAlias+`"]}`, http.StatusCreated)

		response := putConflictTag(t, handler, name, `{"name":"`+name+`","description":"changed","aliases":["`+duplicateAlias+`","`+duplicateAlias+`"]}`, http.StatusConflict)
		assertConflictResponse(t, response)
		assertConflictTag(t, handler, name, "original", []string{oldAlias})
	})

	t.Run("repeated empty aliases are ignored", func(t *testing.T) {
		t.Parallel()
		handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
		name := "empty-aliases-" + uuid.NewV4().String()[:8]

		putConflictTag(t, handler, name, `{"name":"`+name+`","aliases":["",""]}`, http.StatusCreated)
		assertConflictTag(t, handler, name, "", nil)
	})
}

func putConflictTag(t *testing.T, handler http.Handler, name, body string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	response := performRequest(handler, http.MethodPut, "/api/v1/tags/"+name, body, true)
	responseBody := response.Body.String()
	require.Equal(t, wantStatus, response.Code, "put tag %q; body: %s", name, responseBody)
	return response
}

func putConflictCategory(t *testing.T, handler http.Handler, name, body string, wantStatus int) *httptest.ResponseRecorder {
	t.Helper()
	response := performRequest(handler, http.MethodPut, "/api/v1/tagCategories/"+name, body, true)
	responseBody := response.Body.String()
	require.Equal(t, wantStatus, response.Code, "put category %q; body: %s", name, responseBody)
	return response
}

func assertConflictTag(t *testing.T, handler http.Handler, name, description string, aliases []string) {
	t.Helper()
	response := performRequest(handler, http.MethodGet, "/api/v1/tags/"+name, "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "get tag %q; body: %s", name, responseBody)
	var tag types.Tag
	decodeJSON(t, response.Body.Bytes(), &tag)
	assert.Equal(t, name, tag.Name)
	assert.Equal(t, description, tag.Description)
	gotAliases := []string(nil)
	if tag.Aliases != nil {
		gotAliases = *tag.Aliases
	}
	slicesEqual := slices.Equal(gotAliases, aliases)
	assert.True(t, slicesEqual, "tag %q aliases", name)
}

func assertConflictCategory(t *testing.T, handler http.Handler, name, description, color string) {
	t.Helper()
	response := performRequest(handler, http.MethodGet, "/api/v1/tagCategories/"+name, "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "get category %q; body: %s", name, responseBody)
	var category types.TagCategory
	decodeJSON(t, response.Body.Bytes(), &category)
	assert.Equal(t, name, category.Name)
	assert.Equal(t, description, category.Description)
	assert.Equal(t, color, category.Color)
}

func assertConflictResponse(t *testing.T, response *httptest.ResponseRecorder) {
	t.Helper()
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var responseError Error
	decodeJSON(t, response.Body.Bytes(), &responseError)
	assert.NotEmpty(t, responseError.Message, "conflict response message")
}

func TestCreateTag(t *testing.T) {
	t.Parallel()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	tagName := testTagName()

	response := performRequest(handler, http.MethodPut, tagPath(tagName), tagBody(tagName, "created"), true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusCreated, response.Code, "body: %s", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var tag types.Tag
	decodeJSON(t, response.Body.Bytes(), &tag)
	assertTagRequiredFields(t, tag, tagName, "created")
}

func TestGetTag(t *testing.T) {
	t.Parallel()
	handler, tag := createTagFixture(t)

	response := performRequest(handler, http.MethodGet, tagPath(tag.Name), "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var fetched types.Tag
	decodeJSON(t, response.Body.Bytes(), &fetched)
	assert.Equal(t, tag, fetched)
}

func TestListTags(t *testing.T) {
	t.Parallel()
	handler, tag := createTagFixture(t)

	response := performRequest(handler, http.MethodGet, "/api/v1/tags", "", true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var tags TagsResponse
	decodeJSON(t, response.Body.Bytes(), &tags)
	require.NotNil(t, tags.Items)
	require.Len(t, *tags.Items, 1)
	listedTag := (*tags.Items)[0]
	assertTagRequiredFields(t, listedTag, tag.Name, tag.Description)
	require.NotNil(t, listedTag.PostCount)
	assert.Zero(t, *listedTag.PostCount)
}

func TestTagPostCountExcludesSoftDeletedPosts(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	now := time.Now().UTC()
	tagName := testTagName()
	tag, _, err := sqlStore.UpsertTag(t.Context(), tagName, store.TagInput{Name: tagName}, now)
	require.NoError(t, err)
	post := createTestPost(t, sqlStore)
	_, err = sqlStore.UpdatePost(t.Context(), uuid.UUID(post.ID), "", []string{tagName}, now)
	require.NoError(t, err)

	counts, err := sqlStore.GetTagPostCounts(t.Context(), []uuid.UUID{uuid.UUID(tag.ID)})
	require.NoError(t, err)
	assert.Equal(t, 1, counts[uuid.UUID(tag.ID)])

	_, err = sqlStore.SoftDeletePost(t.Context(), uuid.UUID(post.ID), now)
	require.NoError(t, err)
	counts, err = sqlStore.GetTagPostCounts(t.Context(), []uuid.UUID{uuid.UUID(tag.ID)})
	require.NoError(t, err)
	assert.Zero(t, counts[uuid.UUID(tag.ID)])
}

func TestUpdateTag(t *testing.T) {
	t.Parallel()
	handler, tag := createTagFixture(t)

	response := performRequest(handler, http.MethodPut, tagPath(tag.Name), tagBody(tag.Name, "updated"), true)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body: %s", responseBody)
	var updated types.Tag
	decodeJSON(t, response.Body.Bytes(), &updated)
	assertTagRequiredFields(t, updated, tag.Name, "updated")
	assert.Equal(t, tag.CreatedAt, updated.CreatedAt)
	assert.False(t, updated.UpdatedAt.Before(tag.UpdatedAt))
}

func TestDeleteTag(t *testing.T) {
	t.Parallel()
	handler, tag := createTagFixture(t)

	response := performRequest(handler, http.MethodDelete, tagPath(tag.Name), "", true)
	assertEmptyNoContent(t, response.Code, response.Body.String())

	missingResponse := performRequest(handler, http.MethodGet, tagPath(tag.Name), "", true)
	assert.Equal(t, http.StatusNotFound, missingResponse.Code, "body: %s", missingResponse.Body.String())
}

func createTagFixture(t *testing.T) (http.Handler, types.Tag) {
	t.Helper()
	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	tagName := testTagName()
	response := performRequest(handler, http.MethodPut, tagPath(tagName), tagBody(tagName, "created"), true)
	require.Equal(t, http.StatusCreated, response.Code, "body: %s", response.Body.String())
	var tag types.Tag
	decodeJSON(t, response.Body.Bytes(), &tag)
	return handler, tag
}

func testTagName() string {
	return "tag-" + uuid.NewV4().String()[:8]
}

func tagPath(name string) string {
	return "/api/v1/tags/" + name
}

func tagBody(name, description string) string {
	return `{"name":"` + name + `","description":"` + description + `"}`
}

func assertTagRequiredFields(t *testing.T, tag types.Tag, name, description string) {
	t.Helper()
	assert.Equal(t, name, tag.Name)
	assert.Equal(t, description, tag.Description)
	assert.False(t, tag.CreatedAt.IsZero())
	assert.False(t, tag.UpdatedAt.IsZero())
}

func TestTagsPagination(t *testing.T) {
	t.Parallel()

	handler := newTestHandlerForServer(t, NewServer(newTestStore(t), memory.New()))
	want := []string{"tag-00", "tag-01", "tag-02", "tag-03", "tag-04"}
	for _, name := range want {
		response := performRequest(handler, http.MethodPut, "/api/v1/tags/"+name, `{"name":"`+name+`"}`, true)
		responseBody := response.Body.String()
		require.Equalf(t, http.StatusCreated, response.Code, "create tag %q status = %d; body = %q", name, response.Code, responseBody)
	}

	tagNames := traverseNamedPages(t, handler, "/api/v1/tags", "", func(body []byte) ([]string, *string) {
		var response TagsResponse
		decodeJSON(t, body, &response)
		names := make([]string, len(*response.Items))
		for i, item := range *response.Items {
			names[i] = item.Name
		}
		return names, response.Cursor
	})
	assert.Equalf(t, want, tagNames, "traversed tags = %v, want %v", tagNames, want)

	assertStaleNamedCursor(t, handler, "/api/v1/tags", "tag-015", want[2], func(body []byte) string {
		var response TagsResponse
		decodeJSON(t, body, &response)
		return (*response.Items)[0].Name
	})
}
