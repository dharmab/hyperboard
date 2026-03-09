package api

import (
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func postFromModel(model *models.Post) (types.Post, error) {
	post := types.Post{
		ID:           types.ID(model.ID),
		MimeType:     model.MimeType,
		ContentUrl:   model.ContentURL,
		ThumbnailUrl: model.ThumbnailURL,
		Note:         model.Note,
		HasAudio:     model.HasAudio,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}

	// Extract tag names from loaded tags
	tagNames := make([]types.TagName, 0, len(model.R.Tags))
	for _, tag := range model.R.Tags {
		tagNames = append(tagNames, tag.Name)
	}
	post.Tags = tagNames

	return post, nil
}

var sortTerms = map[string]bool{
	types.SortRandom:    true,
	types.SortCreatedAt: true,
	types.SortUpdatedAt: true,
}

func parseSearch(search string) types.PostSearch {
	postSearch := types.PostSearch{
		Tags: []types.TagName{},
	}

	if search == "" {
		return postSearch
	}

	// Split search string by commas and trim whitespace from each term
	for part := range strings.SplitSeq(search, ",") {
		term := strings.TrimSpace(part)
		if term == "" {
			continue
		}
		if sortValue, ok := strings.CutPrefix(term, "sort:"); ok {
			if sortTerms[sortValue] {
				postSearch.Sort = sortValue
			}
		} else if term == "tagged:true" {
			postSearch.Tagged = types.TaggedFilterTrue
		} else if term == "tagged:false" {
			postSearch.Tagged = types.TaggedFilterFalse
		} else if term == types.TagImage {
			postSearch.TypeImage = true
		} else if term == types.TagVideo {
			postSearch.TypeVideo = true
		} else if term == types.TagAudio {
			postSearch.TypeAudio = true
		} else {
			postSearch.Tags = append(postSearch.Tags, term)
		}
	}

	return postSearch
}

type randomCursor struct {
	Seed   int64 `json:"seed"`
	Offset int   `json:"offset"`
}

func encodeRandomCursor(rc randomCursor) string {
	data, _ := json.Marshal(rc)
	return base64.StdEncoding.EncodeToString(data)
}

func decodeRandomCursor(s string, rc *randomCursor) error {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, rc)
}

func (s *Server) GetPosts(w http.ResponseWriter, r *http.Request, params GetPostsParams) {
	ctx := r.Context()

	mods := []bob.Mod[*dialect.SelectQuery]{}

	search := ""
	if params.Search != nil {
		search = *params.Search
	}
	searchParams := parseSearch(search)

	for _, tagName := range searchParams.Tags {
		// Resolve aliases to canonical tag names
		resolved, err := s.resolveAlias(ctx, tagName)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to resolve tag alias")
			return
		}
		mods = append(mods, sm.Where(psql.Raw(
			`EXISTS (
				SELECT 1 FROM posts_tags pt
				JOIN tags t ON pt.tag_id = t.id
				WHERE pt.post_id = posts.id AND t.name = ?
			)`, resolved,
		)))
	}

	// Apply tagged: filter
	switch searchParams.Tagged {
	case types.TaggedFilterTrue:
		mods = append(mods, sm.Where(psql.Raw(
			`EXISTS (
				SELECT 1 FROM posts_tags pt
				JOIN tags t ON pt.tag_id = t.id
				WHERE pt.post_id = posts.id
			)`,
		)))
	case types.TaggedFilterFalse:
		mods = append(mods, sm.Where(psql.Raw(
			`NOT EXISTS (
				SELECT 1 FROM posts_tags pt
				JOIN tags t ON pt.tag_id = t.id
				WHERE pt.post_id = posts.id
			)`,
		)))
	}

	// Apply type: virtual tag filters
	if searchParams.TypeImage {
		mods = append(mods, sm.Where(models.PostColumns.MimeType.Like(psql.Arg("image/%"))))
	}
	if searchParams.TypeVideo {
		mods = append(mods, sm.Where(models.PostColumns.MimeType.Like(psql.Arg("video/%"))))
	}
	if searchParams.TypeAudio {
		mods = append(mods, sm.Where(models.PostColumns.HasAudio.EQ(psql.Arg(true))))
	}

	limit := parseLimit(params.Limit)

	if searchParams.Sort == types.SortRandom {
		currentSeed := time.Now().Unix() / 21600
		offset := 0

		if params.Cursor != nil && *params.Cursor != "" {
			var rc randomCursor
			if err := decodeRandomCursor(*params.Cursor, &rc); err == nil {
				if rc.Seed == currentSeed {
					offset = rc.Offset
				}
				// if seed differs, use currentSeed with offset=0 (window rolled)
			}
		}

		mods = append(mods,
			sm.OrderBy(psql.Raw("md5(posts.id::text || $1::text)", currentSeed)),
			sm.Limit(int64(limit+1)),
			sm.Offset(int64(offset)),
		)

		posts, err := models.Posts.Query(mods...).All(ctx, s.db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
			return
		}

		if err := posts.LoadTags(ctx, s.db); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
			return
		}

		var nextCursor *string
		if len(posts) > limit {
			posts = posts[:limit]
			rc := randomCursor{Seed: currentSeed, Offset: offset + limit}
			encoded := encodeRandomCursor(rc)
			nextCursor = &encoded
		}

		items := make([]types.Post, 0, len(posts))
		for _, post := range posts {
			postResp, err := postFromModel(post)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
				return
			}
			items = append(items, postResp)
		}
		respond(w, http.StatusOK, PostsResponse{Items: &items, Cursor: nextCursor})
		return
	}

	// Determine sort column (default: created_at, newest first)
	sortCol := models.PostColumns.CreatedAt
	if searchParams.Sort == types.SortUpdatedAt {
		sortCol = models.PostColumns.UpdatedAt
	}
	mods = append(mods, sm.OrderBy(sortCol).Desc())

	if params.Cursor != nil && *params.Cursor != "" {
		decodedCursor, err := deobfuscateCursor(*params.Cursor)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		mods = append(mods, sm.Where(sortCol.LT(psql.Arg(decodedCursor))))
	}

	mods = append(mods, sm.Limit(int64(limit+1)))

	posts, err := models.Posts.Query(mods...).All(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve posts")
		return
	}

	if err := posts.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	cursorTimeFn := func() string {
		if searchParams.Sort == types.SortUpdatedAt {
			return posts[limit-1].UpdatedAt.Format("2006-01-02T15:04:05.999999999Z07:00")
		}
		return posts[limit-1].CreatedAt.Format("2006-01-02T15:04:05.999999999Z07:00")
	}
	hasMore, nextCursor := paginate(len(posts), limit, cursorTimeFn)
	if hasMore {
		posts = posts[:limit]
	}

	items := make([]types.Post, 0, len(posts))
	for _, post := range posts {
		postResp, err := postFromModel(post)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
			return
		}
		items = append(items, postResp)
	}
	respond(w, http.StatusOK, PostsResponse{Items: &items, Cursor: nextCursor})
}

func (s *Server) GetPost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	model, err := models.Posts.Query(
		sm.Where(models.PostColumns.ID.EQ(psql.Arg(postID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	if err := model.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	postResp, err := postFromModel(model)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
		return
	}

	respond(w, http.StatusOK, postResp)
}

func (s *Server) UploadPost(w http.ResponseWriter, r *http.Request, params UploadPostParams) {
	ctx := r.Context()

	force := params.Force != nil && *params.Force

	data, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error().Err(err).Msg("failed to read upload request body")
		respondWithError(w, http.StatusInternalServerError, "Failed to read request body: %v", err)
		return
	}

	mimeStr := r.Header.Get("Content-Type")
	if mimeStr == "" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type header is required")
		return
	}
	// Strip any parameters (e.g. "; charset=utf-8") to get a bare MIME type.
	if idx := strings.Index(mimeStr, ";"); idx != -1 {
		mimeStr = strings.TrimSpace(mimeStr[:idx])
	}

	log.Info().Str("mime", mimeStr).Int("size", len(data)).Msg("processing upload")

	var contentData []byte
	var contentMIME string
	var thumbnailData []byte
	var hasAudioVal bool

	if strings.HasPrefix(mimeStr, "image/") {
		contentData, contentMIME, thumbnailData, err = processImage(data, mimeStr)
		if err != nil {
			log.Error().Err(err).Str("mime", mimeStr).Msg("failed to process image")
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process image: %v", err)
			return
		}
	} else if strings.HasPrefix(mimeStr, "video/") {
		contentData = data
		contentMIME = mimeStr
		thumbnailData, err = processVideo(data)
		if err != nil {
			log.Error().Err(err).Str("mime", mimeStr).Msg("failed to process video")
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process video: %v", err)
			return
		}
		hasAudio, probeErr := probeHasAudio(data)
		if probeErr != nil {
			log.Warn().Err(probeErr).Msg("failed to probe audio; assuming no audio")
		}
		hasAudioVal = hasAudio
	} else {
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported media type: %s", mimeStr)
		return
	}

	hash := sha256.Sum256(contentData)
	hashHex := hex.EncodeToString(hash[:])

	existing, err := models.Posts.Query(
		sm.Where(models.PostColumns.Sha256.EQ(psql.Arg(hashHex))),
	).One(ctx, s.db)
	if err == nil {
		respondWithError(w, http.StatusConflict, "Duplicate of existing post %s", existing.ID)
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		log.Error().Err(err).Msg("failed to check for duplicate post")
		respondWithError(w, http.StatusInternalServerError, "Failed to check for duplicate")
		return
	}

	// Compute perceptual hash from the thumbnail.
	var phashVal *sql.Null[int64]
	pHash, phashErr := dhashFromBytes(thumbnailData)
	if phashErr != nil {
		log.Warn().Err(phashErr).Msg("failed to compute perceptual hash")
	} else {
		phashVal = &sql.Null[int64]{V: pHash, Valid: true}

		// Check for visually similar posts (unless force is set).
		if !force && s.similarityThreshold > 0 {
			similar, err := s.findSimilarPosts(ctx, uuid.Nil, pHash, 5)
			if err != nil {
				log.Error().Err(err).Msg("failed to check for similar posts")
			} else if len(similar) > 0 {
				items := make([]types.Post, 0, len(similar))
				for _, p := range similar {
					postResp, convErr := postFromModel(p)
					if convErr != nil {
						continue
					}
					items = append(items, postResp)
				}
				respond(w, http.StatusConflict, SimilarPostsResponse{
					Message: "Similar posts found",
					Similar: items,
				})
				return
			}
		}
	}

	postID, err := uuid.NewV4()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate post ID")
		respondWithError(w, http.StatusInternalServerError, "Failed to generate post ID: %v", err)
		return
	}

	ext := mimeToExt(contentMIME)
	contentKey := fmt.Sprintf("posts/%s/content.%s", postID, ext)
	thumbnailKey := fmt.Sprintf("posts/%s/thumbnail.webp", postID)

	contentURL, err := s.storage.Upload(ctx, contentKey, contentData, contentMIME)
	if err != nil {
		log.Error().Err(err).Str("key", contentKey).Msg("failed to upload content to storage")
		respondWithError(w, http.StatusInternalServerError, "Failed to upload content: %v", err)
		return
	}

	thumbnailURL, err := s.storage.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		log.Error().Err(err).Str("key", thumbnailKey).Msg("failed to upload thumbnail to storage")
		respondWithError(w, http.StatusInternalServerError, "Failed to upload thumbnail: %v", err)
		return
	}

	id := postID
	model, err := models.Posts.Insert(
		&models.PostSetter{
			ID:           &id,
			MimeType:     &contentMIME,
			ContentURL:   &contentURL,
			ThumbnailURL: &thumbnailURL,
			HasAudio:     &hasAudioVal,
			Sha256:       &hashHex,
			Phash:        phashVal,
			CreatedAt:    now(),
			UpdatedAt:    now(),
		},
	).One(ctx, s.db)
	if err != nil {
		log.Error().Err(err).Msg("failed to insert post into database")
		respondWithError(w, http.StatusInternalServerError, "Failed to store post: %v", err)
		return
	}

	model.R.Tags = nil
	postResp, err := postFromModel(model)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
		return
	}

	respond(w, http.StatusCreated, postResp)
}

// mimeToExt returns a file extension for a given MIME type.
func mimeToExt(mime string) string {
	switch mime {
	case "image/webp":
		return "webp"
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "video/mp4":
		return "mp4"
	case "video/webm":
		return "webm"
	case "video/quicktime":
		return "mov"
	default:
		return "bin"
	}
}

func (s *Server) PutPost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	var post types.Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if uuid.UUID(post.ID) != postID {
		respondWithError(w, http.StatusBadRequest, "Post ID mismatch: got %q in body but %q in URL", post.ID, postID)
		return
	}

	existingPost, err := models.Posts.Query(
		sm.Where(models.PostColumns.ID.EQ(psql.Arg(postID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	err = existingPost.Update(ctx, s.db, &models.PostSetter{
		MimeType:     &post.MimeType,
		ContentURL:   &post.ContentUrl,
		ThumbnailURL: &post.ThumbnailUrl,
		Note:         &post.Note,
		UpdatedAt:    now(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	_, err = models.PostsTags.Delete(
		dm.Where(models.PostsTagColumns.PostID.EQ(psql.Arg(postID))),
	).Exec(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update post tags")
		return
	}

	for _, tagName := range post.Tags {
		// Resolve aliases to canonical tag names
		resolvedName, resolveErr := s.resolveAlias(ctx, tagName)
		if resolveErr != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to resolve tag alias")
			return
		}
		tag, err := models.Tags.Query(
			sm.Where(models.TagColumns.Name.EQ(psql.Arg(resolvedName))),
		).One(ctx, s.db)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				tag, err = models.Tags.Insert(
					&models.TagSetter{
						Name:      &resolvedName,
						CreatedAt: now(),
						UpdatedAt: now(),
					},
				).One(ctx, s.db)
				if err != nil {
					respondWithError(w, http.StatusInternalServerError, "Failed to create tag %q", tagName)
					return
				}
			} else {
				respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag")
				return
			}
		}

		err = existingPost.AttachTags(ctx, s.db, tag)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to attach tag")
			return
		}
	}

	if err := existingPost.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	postResp, err := postFromModel(existingPost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
		return
	}

	respond(w, http.StatusOK, postResp)
}

func (s *Server) ReplacePostContent(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	existingPost, err := models.Posts.Query(
		sm.Where(models.PostColumns.ID.EQ(psql.Arg(postID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	mimeStr := r.Header.Get("Content-Type")
	if mimeStr == "" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type header is required")
		return
	}
	if idx := strings.Index(mimeStr, ";"); idx != -1 {
		mimeStr = strings.TrimSpace(mimeStr[:idx])
	}

	var contentData []byte
	var contentMIME string
	var thumbnailData []byte
	var hasAudioVal bool

	if strings.HasPrefix(mimeStr, "image/") {
		contentData, contentMIME, thumbnailData, err = processImage(data, mimeStr)
		if err != nil {
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process image: %v", err)
			return
		}
	} else if strings.HasPrefix(mimeStr, "video/") {
		contentData = data
		contentMIME = mimeStr
		thumbnailData, err = processVideo(data)
		if err != nil {
			respondWithError(w, http.StatusUnprocessableEntity, "Failed to process video: %v", err)
			return
		}
		hasAudio, probeErr := probeHasAudio(data)
		if probeErr != nil {
			log.Warn().Err(probeErr).Msg("failed to probe audio; assuming no audio")
		}
		hasAudioVal = hasAudio
	} else {
		respondWithError(w, http.StatusUnsupportedMediaType, "Unsupported media type: %s", mimeStr)
		return
	}

	ext := mimeToExt(contentMIME)
	contentKey := fmt.Sprintf("posts/%s/content.%s", postID, ext)
	thumbnailKey := fmt.Sprintf("posts/%s/thumbnail.webp", postID)

	contentURL, err := s.storage.Upload(ctx, contentKey, contentData, contentMIME)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload content")
		return
	}

	thumbnailURL, err := s.storage.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload thumbnail")
		return
	}

	hash := sha256.Sum256(contentData)
	hashHex := hex.EncodeToString(hash[:])

	var phashVal *sql.Null[int64]
	pHash, phashErr := dhashFromBytes(thumbnailData)
	if phashErr != nil {
		log.Warn().Err(phashErr).Msg("failed to compute perceptual hash")
	} else {
		phashVal = &sql.Null[int64]{V: pHash, Valid: true}
	}

	err = existingPost.Update(ctx, s.db, &models.PostSetter{
		MimeType:     &contentMIME,
		ContentURL:   &contentURL,
		ThumbnailURL: &thumbnailURL,
		HasAudio:     &hasAudioVal,
		Sha256:       &hashHex,
		Phash:        phashVal,
		UpdatedAt:    now(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	if err := existingPost.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	postResp, err := postFromModel(existingPost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
		return
	}

	respond(w, http.StatusOK, postResp)
}

func (s *Server) ReplacePostThumbnail(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	existingPost, err := models.Posts.Query(
		sm.Where(models.PostColumns.ID.EQ(psql.Arg(postID))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post")
		return
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to read request body")
		return
	}

	mimeStr := r.Header.Get("Content-Type")
	if mimeStr == "" {
		respondWithError(w, http.StatusUnsupportedMediaType, "Content-Type header is required")
		return
	}
	if idx := strings.Index(mimeStr, ";"); idx != -1 {
		mimeStr = strings.TrimSpace(mimeStr[:idx])
	}

	if !strings.HasPrefix(mimeStr, "image/") {
		respondWithError(w, http.StatusUnsupportedMediaType, "Thumbnail must be an image, got: %s", mimeStr)
		return
	}

	_, _, thumbnailData, err := processImage(data, mimeStr)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Failed to process image: %v", err)
		return
	}

	thumbnailKey := fmt.Sprintf("posts/%s/thumbnail.webp", postID)
	thumbnailURL, err := s.storage.Upload(ctx, thumbnailKey, thumbnailData, "image/webp")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to upload thumbnail")
		return
	}

	err = existingPost.Update(ctx, s.db, &models.PostSetter{
		ThumbnailURL: &thumbnailURL,
		UpdatedAt:    now(),
	})
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update post")
		return
	}

	if err := existingPost.LoadTags(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tags")
		return
	}

	postResp, err := postFromModel(existingPost)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert post")
		return
	}

	respond(w, http.StatusOK, postResp)
}

func (s *Server) DeletePost(w http.ResponseWriter, r *http.Request, id Id) {
	ctx := r.Context()

	postID := uuid.UUID(id)

	_, err := models.Posts.Delete(
		dm.Where(models.PostColumns.ID.EQ(psql.Arg(postID))),
	).Exec(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Post not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete post")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
