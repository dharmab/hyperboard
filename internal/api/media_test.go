package api

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/png"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"uuid"

	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/internal/media"
	"github.com/dharmab/hyperboard/internal/storage"
	"github.com/dharmab/hyperboard/internal/storage/memory"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMediaLifecycle(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	originalImage := encodeTestPNG(t, color.RGBA{R: 220, G: 30, B: 40, A: 255})

	created := uploadTestMedia(t, handler, originalImage, "image/png")
	post := created.Post
	assertUploadedPost(t, post, "image/webp", false)
	assert.Nil(t, created.Similar)
	assertPostPersistenceAndRoute(t, sqlStore, handler, post)

	postPath := "/api/v1/posts/" + uuid.UUID(post.ID).String()
	content := assertRoutedMediaResponse(t, handler, postPath+"/content", "image/webp", "")
	thumbnail := assertRoutedMediaResponse(t, handler, postPath+"/thumbnail", "image/webp", "")
	wantDisposition := `attachment; filename="` + uuid.UUID(post.ID).String() + `.webp"`
	download := assertRoutedMediaResponse(t, handler, postPath+"/content/download", "image/webp", wantDisposition)
	assert.Equal(t, content, download)

	replacementImage := encodeTestPNG(t, color.RGBA{R: 20, G: 180, B: 90, A: 255})
	replaced := assertMediaWriteResponse(t, performMediaRequest(handler, http.MethodPut, postPath+"/content", replacementImage, "image/png; charset=binary"), post)
	assert.Equal(t, post.MimeType, replaced.MimeType)
	assert.Equal(t, post.HasAudio, replaced.HasAudio)
	assert.Equal(t, post.ContentUrl, replaced.ContentUrl)
	assert.Equal(t, post.ThumbnailUrl, replaced.ThumbnailUrl)
	assertPostPersistenceAndRoute(t, sqlStore, handler, replaced)
	replacedContent := assertRoutedMediaResponse(t, handler, postPath+"/content", "image/webp", "")
	assert.NotEqual(t, content, replacedContent)

	animatedThumbnail := encodeTestGIF(t, color.RGBA{R: 30, G: 70, B: 230, A: 255})
	thumbnailReplaced := assertMediaWriteResponse(t, performMediaRequest(handler, http.MethodPut, postPath+"/thumbnail", animatedThumbnail, "image/gif"), replaced)
	assertStableContentFields(t, replaced, thumbnailReplaced)
	assertPostPersistenceAndRoute(t, sqlStore, handler, thumbnailReplaced)
	customThumbnail := assertRoutedMediaResponse(t, handler, postPath+"/thumbnail", "image/webp", "")
	assert.NotEqual(t, thumbnail, customThumbnail)

	regenerated := assertMediaWriteResponse(t, performMediaRequest(handler, http.MethodPost, postPath+"/thumbnail", nil, ""), thumbnailReplaced)
	assertStableContentFields(t, thumbnailReplaced, regenerated)
	assertPostPersistenceAndRoute(t, sqlStore, handler, regenerated)
	regeneratedThumbnail := assertRoutedMediaResponse(t, handler, postPath+"/thumbnail", "image/webp", "")
	assert.NotEqual(t, customThumbnail, regeneratedThumbnail)
}

type trackingMediaStore struct {
	storage.MediaStore
	metadataCalls int
	downloadCalls int
	rangeCalls    int
}

func (s *trackingMediaStore) Metadata(ctx context.Context, key string) (*storage.Metadata, error) {
	s.metadataCalls++
	return s.MediaStore.Metadata(ctx, key)
}

func (s *trackingMediaStore) Download(ctx context.Context, key string) (*storage.Media, error) {
	s.downloadCalls++
	return s.MediaStore.Download(ctx, key)
}

func (s *trackingMediaStore) DownloadRange(ctx context.Context, key string, start, end int64) (*storage.Media, error) {
	s.rangeCalls++
	return s.MediaStore.DownloadRange(ctx, key, start, end)
}

func (s *trackingMediaStore) resetCalls() {
	s.metadataCalls = 0
	s.downloadCalls = 0
	s.rangeCalls = 0
}

type failingReadMediaStore struct {
	storage.MediaStore
	metadata    *storage.Metadata
	metadataErr error
	downloadErr error
	rangeErr    error
}

func (s *failingReadMediaStore) Metadata(ctx context.Context, key string) (*storage.Metadata, error) {
	if s.metadataErr != nil {
		return nil, s.metadataErr
	}
	if s.metadata != nil {
		return s.metadata, nil
	}
	return s.MediaStore.Metadata(ctx, key)
}

func (s *failingReadMediaStore) Download(ctx context.Context, key string) (*storage.Media, error) {
	if s.downloadErr != nil {
		return nil, s.downloadErr
	}
	return s.MediaStore.Download(ctx, key)
}

func (s *failingReadMediaStore) DownloadRange(ctx context.Context, key string, start, end int64) (*storage.Media, error) {
	if s.rangeErr != nil {
		return nil, s.rangeErr
	}
	return s.MediaStore.DownloadRange(ctx, key, start, end)
}

func TestMediaReadStorageErrorStatus(t *testing.T) {
	t.Parallel()

	missingErr := fmt.Errorf("wrapped missing object: %w", storage.ErrNotFound)
	dependencyErr := errors.New("object store unavailable")
	for _, tt := range []struct {
		name       string
		store      *failingReadMediaStore
		method     string
		rangeValue string
		wantStatus int
	}{
		{name: "missing download", store: &failingReadMediaStore{MediaStore: memory.New(), downloadErr: missingErr}, method: http.MethodGet, wantStatus: http.StatusNotFound},
		{name: "download dependency failure", store: &failingReadMediaStore{MediaStore: memory.New(), downloadErr: dependencyErr}, method: http.MethodGet, wantStatus: http.StatusServiceUnavailable},
		{name: "missing metadata", store: &failingReadMediaStore{MediaStore: memory.New(), metadataErr: missingErr}, method: http.MethodHead, wantStatus: http.StatusNotFound},
		{name: "metadata dependency failure", store: &failingReadMediaStore{MediaStore: memory.New(), metadataErr: dependencyErr}, method: http.MethodHead, wantStatus: http.StatusServiceUnavailable},
		{name: "missing range download", store: &failingReadMediaStore{MediaStore: memory.New(), metadata: &storage.Metadata{ContentType: "video/mp4", ContentLength: 10}, rangeErr: missingErr}, method: http.MethodGet, rangeValue: "bytes=0-1", wantStatus: http.StatusNotFound},
		{name: "range dependency failure", store: &failingReadMediaStore{MediaStore: memory.New(), metadata: &storage.Metadata{ContentType: "video/mp4", ContentLength: 10}, rangeErr: dependencyErr}, method: http.MethodGet, rangeValue: "bytes=0-1", wantStatus: http.StatusServiceUnavailable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqlStore := newTestStore(t)
			post := createTestPost(t, sqlStore)
			server := NewServer(sqlStore, tt.store)
			request := httptest.NewRequestWithContext(t.Context(), tt.method, "/api/v1/posts/"+post.ID.String()+"/content", nil)
			if tt.rangeValue != "" {
				request.Header.Set("Range", tt.rangeValue)
			}
			response := httptest.NewRecorder()
			if tt.method == http.MethodHead {
				server.HeadPostContent(response, request, types.ID(post.ID))
			} else {
				server.GetPostContent(response, request, types.ID(post.ID))
			}
			assert.Equal(t, tt.wantStatus, response.Code, "body = %q", response.Body.String())
			assertJSONContentType(t, response.Header().Get("Content-Type"))
		})
	}
}

func TestMediaHeadAndRanges(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	memoryStore := memory.New()
	mediaStore := &trackingMediaStore{MediaStore: memoryStore}
	server := NewServer(sqlStore, mediaStore)
	handler := newTestHandlerForServer(t, server)
	created := uploadTestMedia(t, handler, encodeTestPNG(t, color.RGBA{R: 20, G: 80, B: 160, A: 255}), "image/png")
	postID := uuid.UUID(created.Post.ID)
	path := "/api/v1/posts/" + postID.String() + "/content"

	stored, err := memoryStore.Download(t.Context(), storageKeyForContent(postID, created.Post.MimeType))
	require.NoError(t, err)
	content, err := io.ReadAll(stored.Body)
	require.NoError(t, err)
	require.NoError(t, stored.Body.Close())
	require.Greater(t, len(content), 4)

	mediaStore.resetCalls()
	head := performMediaRequest(handler, http.MethodHead, path, nil, "")
	require.Equal(t, http.StatusOK, head.Code, "body = %q", head.Body.String())
	assert.Empty(t, head.Body.Bytes())
	assert.Equal(t, strconv.Itoa(len(content)), head.Header().Get("Content-Length"))
	assert.Equal(t, "bytes", head.Header().Get("Accept-Ranges"))
	assert.Equal(t, 1, mediaStore.metadataCalls)
	assert.Zero(t, mediaStore.downloadCalls)
	assert.Zero(t, mediaStore.rangeCalls)

	mediaStore.resetCalls()
	rangeRequest := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	rangeRequest.SetBasicAuth("any-user", testPassword)
	rangeRequest.Header.Set("Range", "bytes=1-3")
	rangeResponse := httptest.NewRecorder()
	handler.ServeHTTP(rangeResponse, rangeRequest)
	require.Equal(t, http.StatusPartialContent, rangeResponse.Code, "body = %q", rangeResponse.Body.String())
	assert.Equal(t, content[1:4], rangeResponse.Body.Bytes())
	assert.Equal(t, "3", rangeResponse.Header().Get("Content-Length"))
	assert.Equal(t, fmt.Sprintf("bytes 1-3/%d", len(content)), rangeResponse.Header().Get("Content-Range"))
	assert.Equal(t, 1, mediaStore.metadataCalls)
	assert.Zero(t, mediaStore.downloadCalls)
	assert.Equal(t, 1, mediaStore.rangeCalls)

	mediaStore.resetCalls()
	invalidRequest := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
	invalidRequest.SetBasicAuth("any-user", testPassword)
	invalidRequest.Header.Set("Range", fmt.Sprintf("bytes=%d-", len(content)))
	invalidResponse := httptest.NewRecorder()
	handler.ServeHTTP(invalidResponse, invalidRequest)
	require.Equal(t, http.StatusRequestedRangeNotSatisfiable, invalidResponse.Code, "body = %q", invalidResponse.Body.String())
	assert.Equal(t, fmt.Sprintf("bytes */%d", len(content)), invalidResponse.Header().Get("Content-Range"))
	assert.Equal(t, 1, mediaStore.metadataCalls)
	assert.Zero(t, mediaStore.downloadCalls)
	assert.Zero(t, mediaStore.rangeCalls)
}

func TestMediaProcessingErrorStatus(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		err        error
		wantStatus int
	}{
		{name: "invalid client media", err: fmt.Errorf("decode: %w", media.ErrInvalidMedia), wantStatus: http.StatusUnprocessableEntity},
		{name: "server processing failure", err: errors.New("cwebp executable not found"), wantStatus: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			response := httptest.NewRecorder()
			respondToMediaProcessingError(response, tt.err, "Failed to process media")
			assert.Equal(t, tt.wantStatus, response.Code)
			assertJSONContentType(t, response.Header().Get("Content-Type"))
		})
	}
}

func TestMediaBodySizeLimit(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	fixture := encodeTestPNG(t, color.RGBA{R: 100, G: 120, B: 140, A: 255})
	uploadResponse := performMediaRequest(handler, http.MethodPost, "/api/v1/upload", fixture, "image/png")
	uploadResponseBody := uploadResponse.Body.String()
	require.Equal(t, http.StatusCreated, uploadResponse.Code, "body = %q", uploadResponseBody)
	var created CreatedPostResponse
	decodeJSON(t, uploadResponse.Body.Bytes(), &created)
	postPath := "/api/v1/posts/" + uuid.UUID(created.Post.ID).String()

	for _, tt := range []struct {
		name   string
		method string
		target string
	}{
		{name: "upload", method: http.MethodPost, target: "/api/v1/upload"},
		{name: "replace content", method: http.MethodPut, target: postPath + "/content"},
		{name: "replace thumbnail", method: http.MethodPut, target: postPath + "/thumbnail"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			overLimit := performDeclaredSizeMediaRequest(handler, tt.method, tt.target, fixture, maxMediaBody+1)
			overLimitBody := overLimit.Body.String()
			require.Equal(t, http.StatusRequestEntityTooLarge, overLimit.Code, "body = %q", overLimitBody)
			assertJSONContentType(t, overLimit.Header().Get("Content-Type"))

			atLimit := performDeclaredSizeMediaRequest(handler, tt.method, tt.target, []byte("not an image"), maxMediaBody)
			atLimitBody := atLimit.Body.String()
			require.Equal(t, http.StatusUnprocessableEntity, atLimit.Code, "body = %q", atLimitBody)
			assertJSONContentType(t, atLimit.Header().Get("Content-Type"))
		})
	}
}

func TestAnimatedGIFContent(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	original := encodeTestGIF(t, color.RGBA{R: 210, G: 40, B: 30, A: 255})

	created := uploadTestMedia(t, handler, original, "image/gif")
	assertUploadedPost(t, created.Post, "image/gif", false)
	assertPostPersistenceAndRoute(t, sqlStore, handler, created.Post)
	postPath := "/api/v1/posts/" + uuid.UUID(created.Post.ID).String()
	content := assertRoutedMediaResponse(t, handler, postPath+"/content", "image/gif", "")
	assert.Equal(t, original, content)
	assertRoutedMediaResponse(t, handler, postPath+"/thumbnail", "image/webp", "")
}

func TestMP4WithAudioReplacedByImage(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	video := encodeTestMP4WithAudio(t, color.RGBA{R: 40, G: 90, B: 210, A: 255})

	created := uploadTestMedia(t, handler, video, "video/mp4")
	post := created.Post
	assertUploadedPost(t, post, "video/mp4", true)
	assertPostPersistenceAndRoute(t, sqlStore, handler, post)
	postID := uuid.UUID(post.ID)
	postPath := "/api/v1/posts/" + postID.String()
	oldKey := storageKeyForContent(postID, post.MimeType)
	content := assertRoutedMediaResponse(t, handler, postPath+"/content", "video/mp4", "")
	assert.Equal(t, video, content)

	image := encodeTestPNG(t, color.RGBA{R: 20, G: 170, B: 100, A: 255})
	replaced := assertMediaWriteResponse(t, performMediaRequest(handler, http.MethodPut, postPath+"/content", image, "image/png"), post)
	assert.Equal(t, "image/webp", replaced.MimeType)
	assert.False(t, replaced.HasAudio)
	assert.NotEqual(t, post.ContentUrl, replaced.ContentUrl)
	assertPostPersistenceAndRoute(t, sqlStore, handler, replaced)
	assertRoutedMediaResponse(t, handler, postPath+"/content", "image/webp", "")
	_, err := server.mediaStore.Download(t.Context(), oldKey)
	assert.Error(t, err, "superseded MP4 key %q still exists", oldKey)
}

func TestUploadReturnsSimilarPosts(t *testing.T) {
	t.Parallel()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	handler := newTestHandlerForServer(t, server)
	first := uploadTestMedia(t, handler, encodeTestPNG(t, color.RGBA{R: 180, G: 30, B: 30, A: 255}), "image/png")
	second := uploadTestMedia(t, handler, encodeTestPNG(t, color.RGBA{R: 30, G: 30, B: 180, A: 255}), "image/png")
	require.NotNil(t, second.Similar)
	require.Len(t, *second.Similar, 1)
	similar := (*second.Similar)[0]
	assert.Equal(t, first.Post.ID, similar.ID)
	assertPostFields(t, similar)
	assertPostPersistenceAndRoute(t, sqlStore, handler, second.Post)
}

func performMediaRequest(handler http.Handler, method, target string, body []byte, contentType string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(tContext(), method, target, bytes.NewReader(body))
	req.SetBasicAuth("any-user", testPassword)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func performDeclaredSizeMediaRequest(handler http.Handler, method, target string, body []byte, contentLength int64) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(tContext(), method, target, bytes.NewReader(body))
	req.SetBasicAuth("any-user", testPassword)
	req.Header.Set("Content-Type", "image/png")
	req.ContentLength = contentLength
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func tContext() context.Context {
	return context.Background()
}

func assertRoutedMediaResponse(t *testing.T, handler http.Handler, target, contentType, disposition string) []byte {
	t.Helper()
	response := performMediaRequest(handler, http.MethodGet, target, nil, "")
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "GET %s body = %q", target, responseBody)
	responseContentType := response.Header().Get("Content-Type")
	assert.Equal(t, contentType, responseContentType, "GET %s", target)
	responseContentDisposition := response.Header().Get("Content-Disposition")
	assert.Equal(t, disposition, responseContentDisposition, "GET %s", target)
	responseCacheControl := response.Header().Get("Cache-Control")
	assert.Equal(t, "private, max-age=86400", responseCacheControl, "GET %s", target)
	responseContentTypeOptions := response.Header().Get("X-Content-Type-Options")
	assert.Equal(t, "nosniff", responseContentTypeOptions, "GET %s", target)
	mediaBody := response.Body.Bytes()
	require.NotEmpty(t, mediaBody, "GET %s", target)
	return mediaBody
}

func uploadTestMedia(t *testing.T, handler http.Handler, body []byte, contentType string) CreatedPostResponse {
	t.Helper()
	response := performMediaRequest(handler, http.MethodPost, "/api/v1/upload", body, contentType)
	responseBody := response.Body.String()
	require.Equal(t, http.StatusCreated, response.Code, "body = %q", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var created CreatedPostResponse
	decodeJSON(t, response.Body.Bytes(), &created)
	return created
}

func assertUploadedPost(t *testing.T, post types.Post, wantMIME string, wantAudio bool) {
	t.Helper()
	assertPostFields(t, post)
	postID := uuid.UUID(post.ID)
	assert.Equal(t, wantMIME, post.MimeType)
	assert.Equal(t, wantAudio, post.HasAudio)
	wantContentURL := "http://fake-storage/" + storageKeyForContent(postID, wantMIME)
	assert.Equal(t, wantContentURL, post.ContentUrl)
	wantThumbnailURL := "http://fake-storage/" + storageKeyForThumbnail(postID)
	assert.Equal(t, wantThumbnailURL, post.ThumbnailUrl)
	assert.Empty(t, post.Note)
	assert.Empty(t, post.Tags)
	assert.Nil(t, post.TagColors)
	assert.Nil(t, post.CascadingTags)
	assert.Equal(t, post.CreatedAt, post.UpdatedAt)
}

func assertPostFields(t *testing.T, post types.Post) {
	t.Helper()
	nilID := types.ID(uuid.Nil())
	assert.NotEqual(t, nilID, post.ID)
	assert.NotEmpty(t, post.MimeType)
	assert.NotEmpty(t, post.ContentUrl)
	assert.NotEmpty(t, post.ThumbnailUrl)
	assert.False(t, post.CreatedAt.IsZero())
	assert.False(t, post.UpdatedAt.IsZero())
	assert.Empty(t, post.Tags)
}

func assertMediaWriteResponse(t *testing.T, response *httptest.ResponseRecorder, before types.Post) types.Post {
	t.Helper()
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body = %q", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var post types.Post
	decodeJSON(t, response.Body.Bytes(), &post)
	assertPostFields(t, post)
	assert.Equal(t, before.ID, post.ID)
	assert.Equal(t, before.CreatedAt, post.CreatedAt)
	assert.Equal(t, before.Note, post.Note)
	assert.ElementsMatch(t, before.Tags, post.Tags)
	assert.Equal(t, before.TagColors, post.TagColors)
	assert.Equal(t, before.CascadingTags, post.CascadingTags)
	assert.True(t, post.UpdatedAt.After(before.UpdatedAt))
	postID := uuid.UUID(post.ID)
	wantContentURL := "http://fake-storage/" + storageKeyForContent(postID, post.MimeType)
	assert.Equal(t, wantContentURL, post.ContentUrl)
	wantThumbnailURL := "http://fake-storage/" + storageKeyForThumbnail(postID)
	assert.Equal(t, wantThumbnailURL, post.ThumbnailUrl)
	return post
}

func assertStableContentFields(t *testing.T, before, after types.Post) {
	t.Helper()
	assert.Equal(t, before.MimeType, after.MimeType)
	assert.Equal(t, before.ContentUrl, after.ContentUrl)
	assert.Equal(t, before.ThumbnailUrl, after.ThumbnailUrl)
	assert.Equal(t, before.HasAudio, after.HasAudio)
}

func assertPostPersistenceAndRoute(t *testing.T, sqlStore store.SQLStore, handler http.Handler, want types.Post) {
	t.Helper()
	postID := uuid.UUID(want.ID)
	model, err := sqlStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get persisted post %s", postID)
	postIDString := postID.String()
	modelIDString := model.ID.String()
	assert.Equal(t, postIDString, modelIDString)
	assert.Equal(t, want.MimeType, model.MimeType)
	assert.Equal(t, want.ContentUrl, model.ContentURL)
	assert.Equal(t, want.ThumbnailUrl, model.ThumbnailURL)
	assert.Equal(t, want.Note, model.Note)
	assert.Equal(t, want.HasAudio, model.HasAudio)
	assert.True(t, model.FileSize.Valid)
	assert.WithinDuration(t, want.CreatedAt, model.CreatedAt, 0)
	assert.WithinDuration(t, want.UpdatedAt, model.UpdatedAt, 0)
	response := performMediaRequest(handler, http.MethodGet, "/api/v1/posts/"+postID.String(), nil, "")
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "body = %q", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	var routed types.Post
	decodeJSON(t, response.Body.Bytes(), &routed)
	assert.Equal(t, want, routed)

	contentResponse := performMediaRequest(handler, http.MethodGet, "/api/v1/posts/"+postID.String()+"/content", nil, "")
	require.Equal(t, http.StatusOK, contentResponse.Code, "body = %q", contentResponse.Body.String())
	assert.Equal(t, int64(contentResponse.Body.Len()), model.FileSize.V)
}

func encodeTestGIF(t *testing.T, fill color.Color) []byte {
	t.Helper()
	palette := color.Palette{fill, color.Black}
	frames := []*image.Paletted{
		image.NewPaletted(image.Rect(0, 0, 8, 6), palette),
		image.NewPaletted(image.Rect(0, 0, 8, 6), palette),
	}
	for i := range frames[1].Pix {
		frames[1].Pix[i] = 1
	}
	var data bytes.Buffer
	err := gif.EncodeAll(&data, &gif.GIF{Image: frames, Delay: []int{5, 5}})
	require.NoError(t, err, "encode GIF fixture")
	return data.Bytes()
}

func encodeTestMP4WithAudio(t *testing.T, fill color.Color) []byte {
	t.Helper()
	rgba := color.RGBAModel.Convert(fill).(color.RGBA)
	colorValue := fmt.Sprintf("0x%02x%02x%02x", rgba.R, rgba.G, rgba.B)
	path := t.TempDir() + "/fixture.mp4"
	cmd := exec.CommandContext(t.Context(), "ffmpeg", "-v", "error", "-f", "lavfi", "-i", "color=c="+colorValue+":s=16x16:d=2:r=2", "-f", "lavfi", "-i", "sine=frequency=1000:duration=2", "-c:v", "mpeg4", "-pix_fmt", "yuv420p", "-c:a", "aac", "-shortest", "-y", path) //nolint:gosec // test arguments are generated locally
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "encode MP4 fixture with audio: %s", output)
	data, err := os.ReadFile(path)
	require.NoError(t, err, "read MP4 fixture")
	return data
}

func encodeTestPNG(t *testing.T, fill color.Color) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 8, 6))
	for y := range 6 {
		for x := range 8 {
			img.Set(x, y, fill)
		}
	}
	var data bytes.Buffer
	err := png.Encode(&data, img)
	require.NoError(t, err, "encode PNG fixture")
	return data.Bytes()
}

var errUnexpectedMediaBodyRead = errors.New("media request body was read")

type readRejectingBody struct {
	read bool
}

func (b *readRejectingBody) Read([]byte) (int, error) {
	b.read = true
	return 0, errUnexpectedMediaBodyRead
}

func (*readRejectingBody) Close() error { return nil }

type mutationCountingMediaStore struct {
	storage.MediaStore
	uploads int
	deletes int
}

func (s *mutationCountingMediaStore) Upload(ctx context.Context, key string, data []byte, contentType string) (string, error) {
	s.uploads++
	return s.MediaStore.Upload(ctx, key, data, contentType)
}

func (s *mutationCountingMediaStore) Delete(ctx context.Context, key string) error {
	s.deletes++
	return s.MediaStore.Delete(ctx, key)
}

func TestUnsupportedMediaTypeDoesNotReadRequestBody(t *testing.T) {
	t.Parallel()

	postID := uuid.NewV4()
	for _, tt := range []struct {
		name        string
		contentType string
		invoke      func(*Server, http.ResponseWriter, *http.Request)
	}{
		{name: "upload missing content type", invoke: func(server *Server, w http.ResponseWriter, r *http.Request) { server.UploadPost(w, r) }},
		{name: "upload unsupported content type", contentType: "text/plain", invoke: func(server *Server, w http.ResponseWriter, r *http.Request) { server.UploadPost(w, r) }},
		{name: "content replacement missing content type", invoke: func(server *Server, w http.ResponseWriter, r *http.Request) {
			server.ReplacePostContent(w, r, Id(postID))
		}},
		{name: "content replacement unsupported content type", contentType: "text/plain", invoke: func(server *Server, w http.ResponseWriter, r *http.Request) {
			server.ReplacePostContent(w, r, Id(postID))
		}},
		{name: "thumbnail replacement missing content type", invoke: func(server *Server, w http.ResponseWriter, r *http.Request) {
			server.ReplacePostThumbnail(w, r, Id(postID))
		}},
		{name: "thumbnail replacement unsupported content type", contentType: "video/mp4", invoke: func(server *Server, w http.ResponseWriter, r *http.Request) {
			server.ReplacePostThumbnail(w, r, Id(postID))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			body := &readRejectingBody{}
			request := httptest.NewRequestWithContext(t.Context(), http.MethodPut, "/", io.NopCloser(body))
			request.Body = body
			if tt.contentType != "" {
				request.Header.Set("Content-Type", tt.contentType)
			}
			response := httptest.NewRecorder()

			tt.invoke(&Server{}, response, request)

			require.Equal(t, http.StatusUnsupportedMediaType, response.Code, "body = %q", response.Body.String())
			assert.False(t, body.read, "unsupported request body was consumed")
		})
	}
}

func TestMalformedMediaUploadsDoNotMutateState(t *testing.T) {
	t.Parallel()

	pngBody := encodeTestPNG(t, color.RGBA{R: 31, G: 91, B: 151, A: 255})
	for _, tt := range []struct {
		name        string
		body        []byte
		contentType string
		wantStatus  int
	}{
		{name: "empty body", body: nil, contentType: "image/png", wantStatus: http.StatusUnprocessableEntity},
		{name: "corrupt image", body: []byte("not a png"), contentType: "image/png", wantStatus: http.StatusUnprocessableEntity},
		{name: "corrupt GIF", body: []byte("GIF89a truncated"), contentType: "image/gif", wantStatus: http.StatusUnprocessableEntity},
		{name: "corrupt video", body: []byte("not an mp4"), contentType: "video/mp4", wantStatus: http.StatusUnprocessableEntity},
		{name: "declared MIME does not match body", body: pngBody, contentType: "image/gif", wantStatus: http.StatusUnprocessableEntity},
		{name: "unsupported content type", body: []byte("plain text"), contentType: "text/plain", wantStatus: http.StatusUnsupportedMediaType},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqlStore := newTestStore(t)
			server := NewServer(sqlStore, memory.New())
			mediaStore := &mutationCountingMediaStore{MediaStore: memory.New()}
			server.mediaStore = mediaStore
			handler := newTestHandlerForServer(t, server)

			response := performMediaRequest(handler, http.MethodPost, "/api/v1/upload", tt.body, tt.contentType)
			responseBody := response.Body.String()
			require.Equal(t, tt.wantStatus, response.Code, "body = %q", responseBody)
			assertJSONContentType(t, response.Header().Get("Content-Type"))
			posts, _, err := sqlStore.ListPosts(t.Context(), store.ListPostsParams{Limit: 1})
			require.NoError(t, err, "list posts after rejected upload")
			assert.Empty(t, posts)
			assert.Zero(t, mediaStore.uploads)
			assert.Zero(t, mediaStore.deletes)
		})
	}
}

func TestExactDuplicateUploadReturnsConflictWithoutMutation(t *testing.T) {
	t.Parallel()

	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	mediaStore := &mutationCountingMediaStore{MediaStore: memory.New()}
	server.mediaStore = mediaStore
	handler := newTestHandlerForServer(t, server)
	body := encodeTestPNG(t, color.RGBA{R: 42, G: 102, B: 162, A: 255})
	created := uploadTestMedia(t, handler, body, "image/png")
	postID := uuid.UUID(created.Post.ID)
	contentKey := storageKeyForContent(postID, created.Post.MimeType)
	thumbnailKey := storageKeyForThumbnail(postID)
	contentBefore := readStoredMedia(t, mediaStore, contentKey)
	thumbnailBefore := readStoredMedia(t, mediaStore, thumbnailKey)
	uploadsBefore, deletesBefore := mediaStore.uploads, mediaStore.deletes

	response := performMediaRequest(handler, http.MethodPost, "/api/v1/upload", body, "image/png")
	responseBody := response.Body.String()
	require.Equal(t, http.StatusConflict, response.Code, "body = %q", responseBody)
	assertJSONContentType(t, response.Header().Get("Content-Type"))
	posts, _, err := sqlStore.ListPosts(t.Context(), store.ListPostsParams{Limit: 2})
	require.NoError(t, err, "list posts after duplicate upload")
	require.Len(t, posts, 1)
	postIDString := postID.String()
	modelIDString := posts[0].ID.String()
	assert.Equal(t, postIDString, modelIDString)
	assert.Equal(t, uploadsBefore, mediaStore.uploads)
	assert.Equal(t, deletesBefore, mediaStore.deletes)
	assertStoredMedia(t, mediaStore, contentKey, contentBefore)
	assertStoredMedia(t, mediaStore, thumbnailKey, thumbnailBefore)
}

func TestMalformedMediaReplacementsPreservePostAndObjects(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name        string
		suffix      string
		body        []byte
		contentType string
		wantStatus  int
	}{
		{name: "empty content", suffix: "/content", body: nil, contentType: "image/png", wantStatus: http.StatusUnprocessableEntity},
		{name: "corrupt image content", suffix: "/content", body: []byte("broken"), contentType: "image/png", wantStatus: http.StatusUnprocessableEntity},
		{name: "corrupt GIF content", suffix: "/content", body: []byte("GIF89a broken"), contentType: "image/gif", wantStatus: http.StatusUnprocessableEntity},
		{name: "corrupt video content", suffix: "/content", body: []byte("broken"), contentType: "video/mp4", wantStatus: http.StatusUnprocessableEntity},
		{name: "mismatched content", suffix: "/content", body: encodeTestPNG(t, color.RGBA{R: 9, G: 19, B: 29, A: 255}), contentType: "image/gif", wantStatus: http.StatusUnprocessableEntity},
		{name: "unsupported content", suffix: "/content", body: []byte("text"), contentType: "text/plain", wantStatus: http.StatusUnsupportedMediaType},
		{name: "empty thumbnail", suffix: "/thumbnail", body: nil, contentType: "image/png", wantStatus: http.StatusUnprocessableEntity},
		{name: "corrupt thumbnail", suffix: "/thumbnail", body: []byte("broken"), contentType: "image/png", wantStatus: http.StatusUnprocessableEntity},
		{name: "mismatched thumbnail", suffix: "/thumbnail", body: encodeTestPNG(t, color.RGBA{R: 10, G: 20, B: 30, A: 255}), contentType: "image/gif", wantStatus: http.StatusUnprocessableEntity},
		{name: "unsupported thumbnail", suffix: "/thumbnail", body: []byte("video"), contentType: "video/mp4", wantStatus: http.StatusUnsupportedMediaType},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqlStore, mediaStore, handler, postID := newMediaErrorFixture(t)
			assertRejectedMediaWritePreservesState(t, sqlStore, mediaStore, handler, postID, tt.suffix, tt.body, tt.contentType, tt.wantStatus)
		})
	}
}

func TestMissingPostMediaWritesDoNotTouchStorage(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		method string
		suffix string
		body   []byte
		mime   string
	}{
		{name: "replace content", method: http.MethodPut, suffix: "/content", body: encodeTestPNG(t, color.Black), mime: "image/png"},
		{name: "replace thumbnail", method: http.MethodPut, suffix: "/thumbnail", body: encodeTestPNG(t, color.White), mime: "image/png"},
		{name: "regenerate thumbnail", method: http.MethodPost, suffix: "/thumbnail"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sqlStore := newTestStore(t)
			server := NewServer(sqlStore, memory.New())
			mediaStore := &mutationCountingMediaStore{MediaStore: memory.New()}
			server.mediaStore = mediaStore
			handler := newTestHandlerForServer(t, server)
			response := performMediaRequest(handler, tt.method, "/api/v1/posts/"+uuid.NewV4().String()+tt.suffix, tt.body, tt.mime)
			responseBody := response.Body.String()
			require.Equal(t, http.StatusNotFound, response.Code, "body = %q", responseBody)
			assert.Zero(t, mediaStore.uploads)
			assert.Zero(t, mediaStore.deletes)
		})
	}
}

func TestRegenerateThumbnailWithMissingContentPreservesState(t *testing.T) {
	t.Parallel()

	sqlStore, mediaStore, handler, postID := newMediaErrorFixture(t)
	postBefore, err := sqlStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get fixture post")
	contentKey := storageKeyForContent(postID, postBefore.MimeType)
	thumbnailKey := storageKeyForThumbnail(postID)
	require.NoError(t, mediaStore.Delete(t.Context(), contentKey), "remove fixture content")
	thumbnailBefore := readStoredMedia(t, mediaStore, thumbnailKey)
	uploadsBefore := mediaStore.uploads

	response := performMediaRequest(handler, http.MethodPost, "/api/v1/posts/"+postID.String()+"/thumbnail", nil, "")
	require.Equal(t, http.StatusInternalServerError, response.Code, "body = %q", response.Body.String())
	postAfter, err := sqlStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get post after failure")
	assert.Equal(t, postBefore, postAfter)
	assertStoredMedia(t, mediaStore, thumbnailKey, thumbnailBefore)
	assertMediaMissingFromStore(t, mediaStore, contentKey)
	assert.Equal(t, uploadsBefore, mediaStore.uploads)
}

func newMediaErrorFixture(t *testing.T) (store.SQLStore, *mutationCountingMediaStore, http.Handler, uuid.UUID) {
	t.Helper()
	sqlStore := newTestStore(t)
	server := NewServer(sqlStore, memory.New())
	mediaStore := &mutationCountingMediaStore{MediaStore: memory.New()}
	server.mediaStore = mediaStore
	handler := newTestHandlerForServer(t, server)
	created := uploadTestMedia(t, handler, encodeTestPNG(t, color.RGBA{R: 81, G: 121, B: 161, A: 255}), "image/png")
	return sqlStore, mediaStore, handler, uuid.UUID(created.Post.ID)
}

func assertRejectedMediaWritePreservesState(t *testing.T, sqlStore store.SQLStore, mediaStore *mutationCountingMediaStore, handler http.Handler, postID uuid.UUID, suffix string, body []byte, contentType string, wantStatus int) {
	t.Helper()
	postBefore, err := sqlStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get fixture post")
	contentKey := storageKeyForContent(postID, postBefore.MimeType)
	thumbnailKey := storageKeyForThumbnail(postID)
	contentBefore := readStoredMedia(t, mediaStore, contentKey)
	thumbnailBefore := readStoredMedia(t, mediaStore, thumbnailKey)
	uploadsBefore, deletesBefore := mediaStore.uploads, mediaStore.deletes

	response := performMediaRequest(handler, http.MethodPut, "/api/v1/posts/"+postID.String()+suffix, body, contentType)
	responseBody := response.Body.String()
	require.Equal(t, wantStatus, response.Code, "body = %q", responseBody)
	postAfter, err := sqlStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get post after rejected replacement")
	assert.Equal(t, postBefore, postAfter)
	assertStoredMedia(t, mediaStore, contentKey, contentBefore)
	assertStoredMedia(t, mediaStore, thumbnailKey, thumbnailBefore)
	assert.Equal(t, uploadsBefore, mediaStore.uploads)
	assert.Equal(t, deletesBefore, mediaStore.deletes)
}

func TestReplaceContentPreservesMediaAfterUploadFailure(t *testing.T) {
	t.Parallel()

	mediaStore := &faultInjectingMediaStore{MediaStore: memory.New()}
	baseStore := newTestStore(t)
	handler := newTestHandlerForServer(t, NewServer(baseStore, mediaStore))
	uploadResponse := performMediaRequest(handler, http.MethodPost, "/api/v1/upload", encodeTestPNG(t, color.RGBA{R: 73, G: 149, B: 211, A: 255}), "image/png")
	uploadResponseBody := uploadResponse.Body.String()
	require.Equal(t, http.StatusCreated, uploadResponse.Code, "response body: %s", uploadResponseBody)
	var created CreatedPostResponse
	decodeJSON(t, uploadResponse.Body.Bytes(), &created)
	postID := uuid.UUID(created.Post.ID)
	contentKey := storageKeyForContent(postID, created.Post.MimeType)
	thumbnailKey := storageKeyForThumbnail(postID)
	originalContent := readStoredMedia(t, mediaStore, contentKey)
	originalThumbnail := readStoredMedia(t, mediaStore, thumbnailKey)
	originalPost, err := baseStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get original post")

	mediaStore.failUploadAt = mediaStore.uploads + 1
	response := performMediaRequest(handler, http.MethodPut, "/api/v1/posts/"+postID.String()+"/content", encodeTestPNG(t, color.RGBA{R: 191, G: 81, B: 31, A: 255}), "image/png")
	responseBody := response.Body.String()
	require.Equal(t, http.StatusInternalServerError, response.Code, "response body: %s", responseBody)
	assertStoredMedia(t, mediaStore, contentKey, originalContent)
	assertStoredMedia(t, mediaStore, thumbnailKey, originalThumbnail)
	post, err := baseStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get post after failed replacement")
	assert.Equal(t, originalPost.MimeType, post.MimeType, "post MIME type after upload failure")
	assert.Equal(t, originalPost.ContentURL, post.ContentURL, "post content URL after upload failure")
	assert.Equal(t, originalPost.ThumbnailURL, post.ThumbnailURL, "post thumbnail URL after upload failure")
	assert.Equal(t, originalPost.SHA256, post.SHA256, "post SHA-256 after upload failure")
	assert.Equal(t, originalPost.UpdatedAt, post.UpdatedAt, "post update time after upload failure")
}

func TestReplaceContentChangesStorageKeyAfterMIMEChange(t *testing.T) {
	t.Parallel()

	mediaStore := &faultInjectingMediaStore{MediaStore: memory.New()}
	baseStore := newTestStore(t)
	handler := newTestHandlerForServer(t, NewServer(baseStore, mediaStore))
	uploadResponse := performMediaRequest(handler, http.MethodPost, "/api/v1/upload", encodeTestPNG(t, color.RGBA{R: 74, G: 150, B: 212, A: 255}), "image/png")
	uploadResponseBody := uploadResponse.Body.String()
	require.Equal(t, http.StatusCreated, uploadResponse.Code, "response body: %s", uploadResponseBody)
	var created CreatedPostResponse
	decodeJSON(t, uploadResponse.Body.Bytes(), &created)
	postID := uuid.UUID(created.Post.ID)
	oldContentKey := storageKeyForContent(postID, created.Post.MimeType)

	replacement := encodeTestGIF(t, color.RGBA{R: 21, G: 171, B: 101, A: 255})
	response := performMediaRequest(handler, http.MethodPut, "/api/v1/posts/"+postID.String()+"/content", replacement, "image/gif")
	responseBody := response.Body.String()
	require.Equal(t, http.StatusOK, response.Code, "response body: %s", responseBody)
	var replaced types.Post
	decodeJSON(t, response.Body.Bytes(), &replaced)
	assert.Equal(t, "image/gif", replaced.MimeType, "replacement response MIME type")
	newContentKey := storageKeyForContent(postID, replaced.MimeType)
	newContent := readStoredMedia(t, mediaStore, newContentKey)
	assert.Equal(t, replacement, newContent.data, "replacement content for key %q", newContentKey)
	assert.Equal(t, "image/gif", newContent.contentType, "replacement content type for key %q", newContentKey)
	_, err := mediaStore.Download(t.Context(), oldContentKey)
	require.Error(t, err, "download superseded content key %q", oldContentKey)
	assert.Contains(t, mediaStore.deletedKeys, oldContentKey, "deleted media keys")
	post, err := baseStore.GetPost(t.Context(), postID)
	require.NoError(t, err, "get replaced post")
	assert.Equal(t, "image/gif", post.MimeType, "persisted MIME type")
	assert.Equal(t, replaced.ContentUrl, post.ContentURL, "persisted content URL")
	assert.Equal(t, replaced.ThumbnailUrl, post.ThumbnailURL, "persisted thumbnail URL")
	thumbnail := readStoredMedia(t, mediaStore, storageKeyForThumbnail(postID))
	assert.NotEmpty(t, thumbnail.data, "replacement thumbnail data")
	assert.Equal(t, "image/webp", thumbnail.contentType, "replacement thumbnail content type")
}

func TestThumbnailWritesPreserveStateAfterUploadFailure(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name        string
		method      string
		body        func(*testing.T) []byte
		contentType string
	}{
		{name: "replacement", method: http.MethodPut, body: func(t *testing.T) []byte {
			t.Helper()
			return encodeTestPNG(t, color.RGBA{R: 191, G: 81, B: 31, A: 255})
		}, contentType: "image/png"},
		{name: "regeneration", method: http.MethodPost, body: func(*testing.T) []byte { return nil }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mediaStore := &faultInjectingMediaStore{MediaStore: memory.New()}
			baseStore := newTestStore(t)
			handler := newTestHandlerForServer(t, NewServer(baseStore, mediaStore))
			uploadResponse := performMediaRequest(handler, http.MethodPost, "/api/v1/upload", encodeTestPNG(t, color.RGBA{R: 70, G: 146, B: 208, A: 255}), "image/png")
			uploadResponseBody := uploadResponse.Body.String()
			require.Equal(t, http.StatusCreated, uploadResponse.Code, "response body: %s", uploadResponseBody)
			var created CreatedPostResponse
			decodeJSON(t, uploadResponse.Body.Bytes(), &created)
			postID := uuid.UUID(created.Post.ID)
			contentKey := storageKeyForContent(postID, created.Post.MimeType)
			thumbnailKey := storageKeyForThumbnail(postID)
			content := readStoredMedia(t, mediaStore, contentKey)
			thumbnail := readStoredMedia(t, mediaStore, thumbnailKey)
			originalPost, err := baseStore.GetPost(t.Context(), postID)
			require.NoError(t, err, "get original post")

			mediaStore.failUploadAt = mediaStore.uploads + 1
			response := performMediaRequest(handler, tt.method, "/api/v1/posts/"+postID.String()+"/thumbnail", tt.body(t), tt.contentType)
			responseBody := response.Body.String()
			require.Equal(t, http.StatusInternalServerError, response.Code, "response body: %s", responseBody)
			assertJSONContentType(t, response.Header().Get("Content-Type"))
			assertStoredMedia(t, mediaStore, contentKey, content)
			assertStoredMedia(t, mediaStore, thumbnailKey, thumbnail)
			post, err := baseStore.GetPost(t.Context(), postID)
			require.NoError(t, err, "get post after upload failure")
			assert.Equal(t, originalPost.ThumbnailURL, post.ThumbnailURL, "post thumbnail URL after upload failure")
			assert.Equal(t, originalPost.UpdatedAt, post.UpdatedAt, "post update time after upload failure")
		})
	}
}

type storedMedia struct {
	data        []byte
	contentType string
}

func readStoredMedia(t *testing.T, mediaStore storage.MediaStore, key string) storedMedia {
	t.Helper()
	media, err := mediaStore.Download(t.Context(), key)
	require.NoError(t, err, "download stored media %q", key)
	t.Cleanup(func() {
		assert.NoError(t, media.Body.Close(), "close stored media %q", key)
	})
	data, err := io.ReadAll(media.Body)
	require.NoError(t, err, "read stored media %q", key)
	return storedMedia{data: data, contentType: media.ContentType}
}

func assertStoredMedia(t *testing.T, mediaStore storage.MediaStore, key string, want storedMedia) {
	t.Helper()
	storedMedia := readStoredMedia(t, mediaStore, key)
	assert.Equal(t, want.contentType, storedMedia.contentType, "stored media %q content type", key)
	assert.Equal(t, want.data, storedMedia.data, "stored media %q data", key)
}

func assertMediaMissingFromStore(t *testing.T, mediaStore storage.MediaStore, key string) {
	t.Helper()
	_, err := mediaStore.Download(t.Context(), key)
	assert.Error(t, err, "stored media %q", key)
}
