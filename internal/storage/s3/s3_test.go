package s3

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCreatesBucketWithRegionConstraint(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name           string
		region         string
		wantConstraint string
	}{
		{name: "empty region", region: ""},
		{name: "us-east-1", region: "us-east-1"},
		{name: "AWS regional bucket", region: "eu-west-1", wantConstraint: "eu-west-1"},
		{name: "S3-compatible custom region", region: "auto", wantConstraint: "auto"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var createBody string
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.Method {
				case http.MethodHead:
					w.WriteHeader(http.StatusNotFound)
				case http.MethodPut:
					body, err := io.ReadAll(r.Body)
					if err != nil {
						t.Errorf("read create-bucket request: %v", err)
						http.Error(w, "failed to read request", http.StatusInternalServerError)
						return
					}
					createBody = string(body)
					w.WriteHeader(http.StatusOK)
				default:
					http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
				}
			}))
			defer server.Close()

			_, err := New(t.Context(), server.URL, "media", "access", "secret", tt.region, true)
			require.NoError(t, err)
			if tt.wantConstraint == "" {
				assert.Empty(t, createBody)
			} else {
				assert.Contains(t, createBody, "<LocationConstraint>"+tt.wantConstraint+"</LocationConstraint>")
			}
		})
	}
}

func TestNewAcceptsConcurrentBucketCreationAfterVerification(t *testing.T) {
	t.Parallel()

	headCalls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodHead:
			headCalls++
			if headCalls == 1 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
		case http.MethodPut:
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusConflict)
			_, _ = io.WriteString(w, `<Error><Code>BucketAlreadyOwnedByYou</Code><Message>already created</Message></Error>`)
		default:
			http.Error(w, "unexpected method", http.StatusMethodNotAllowed)
		}
	}))
	defer server.Close()

	_, err := New(t.Context(), server.URL, "media", "access", "secret", "us-west-2", true)
	require.NoError(t, err)
	assert.Equal(t, 2, headCalls)
}

func TestMetadataDoesNotClassifyMissingBucketAsMissingObject(t *testing.T) {
	t.Parallel()

	bucketChecks := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && r.URL.Path == "/media" {
			bucketChecks++
			if bucketChecks == 1 {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mediaStore, err := New(t.Context(), server.URL, "media", "access", "secret", "us-east-1", true)
	require.NoError(t, err)
	_, err = mediaStore.Metadata(t.Context(), "missing")
	require.Error(t, err)
	require.NotErrorIs(t, err, storage.ErrNotFound)
	assert.GreaterOrEqual(t, bucketChecks, 2)
}

func TestObjectErrorsClassifyOnlyMissingObjects(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead && r.URL.Path == "/media" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/missing") {
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusNotFound)
			_, _ = io.WriteString(w, `<Error><Code>NoSuchKey</Code><Message>missing</Message></Error>`)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, `<Error><Code>ServiceUnavailable</Code><Message>try later</Message></Error>`)
	}))
	defer server.Close()

	mediaStore, err := New(t.Context(), server.URL, "media", "access", "secret", "us-east-1", true)
	require.NoError(t, err)

	_, metadataErr := mediaStore.Metadata(t.Context(), "missing")
	require.ErrorIs(t, metadataErr, storage.ErrNotFound)
	_, downloadErr := mediaStore.Download(t.Context(), "missing")
	require.ErrorIs(t, downloadErr, storage.ErrNotFound)
	_, rangeErr := mediaStore.DownloadRange(t.Context(), "missing", 0, 1)
	require.ErrorIs(t, rangeErr, storage.ErrNotFound)

	_, dependencyMetadataErr := mediaStore.Metadata(t.Context(), "unavailable")
	require.Error(t, dependencyMetadataErr)
	require.NotErrorIs(t, dependencyMetadataErr, storage.ErrNotFound)
	_, dependencyDownloadErr := mediaStore.Download(t.Context(), "unavailable")
	require.Error(t, dependencyDownloadErr)
	require.NotErrorIs(t, dependencyDownloadErr, storage.ErrNotFound)
}
