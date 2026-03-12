package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
)

func TestHandleTags(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()
	tags := []types.Tag{{Name: "test-tag", CreatedAt: now, UpdatedAt: now}}
	cats := []types.TagCategory{}

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/tags") {
			jsonResponse(w, http.StatusOK, client.TagsResponse{Items: &tags})
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/v1/tagCategories") {
			jsonResponse(w, http.StatusOK, client.TagCategoriesResponse{Items: &cats})
			return
		}
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/tags", nil)
	w := httptest.NewRecorder()
	app.handleTags(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "test-tag") {
		t.Error("expected tag name in response body")
	}
}

func TestHandleTagConvertToAlias(t *testing.T) {
	t.Parallel()
	now := time.Now().UTC()

	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/convert-to-alias") {
			jsonResponse(w, http.StatusOK, types.Tag{Name: "target-tag", CreatedAt: now, UpdatedAt: now})
			return
		}
		http.NotFound(w, r)
	}))

	form := strings.NewReader("target=target-tag")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/tags/source-tag/convert-to-alias", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("name", "source-tag")
	w := httptest.NewRecorder()
	app.handleTagConvertToAlias(w, req)

	if w.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusSeeOther, w.Body.String())
	}
	loc := w.Header().Get("Location")
	if loc != "/tags/target-tag" {
		t.Errorf("redirect location = %q, want /tags/target-tag", loc)
	}
}

func TestHandleTagConvertToAlias_EmptyTarget(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/tags/source-tag/convert-to-alias", nil)
	req.SetPathValue("name", "source-tag")
	w := httptest.NewRecorder()
	app.handleTagConvertToAlias(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleTagConvertToAlias_SameTarget(t *testing.T) {
	t.Parallel()
	app := newTestApp(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	form := strings.NewReader("target=source-tag")
	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/tags/source-tag/convert-to-alias", form)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("name", "source-tag")
	w := httptest.NewRecorder()
	app.handleTagConvertToAlias(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
