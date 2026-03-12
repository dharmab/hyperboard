package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/db/store"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/rs/zerolog"
)

func tagFromModel(model *models.Tag) types.Tag {
	tag := types.Tag{
		Name:        model.Name,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	// Load the tag category if present
	if model.TagCategoryID.Valid {
		if model.TagCategory != nil {
			tag.Category = &model.TagCategory.Name
		}
	}

	return tag
}

func (s *Server) GetTags(w http.ResponseWriter, r *http.Request, params GetTagsParams) {
	ctx := r.Context()

	var decodedCursor *string
	if params.Cursor != nil && *params.Cursor != "" {
		decoded, err := deobfuscateCursor(*params.Cursor)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		decodedCursor = &decoded
	}

	limit := parseLimit(params.Limit)

	tags, hasMore, err := s.sqlStore.ListTags(ctx, decodedCursor, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tags")
		return
	}

	// Collect tag IDs for the current page
	tagIDs := make([]uuid.UUID, len(tags))
	for i := range tags {
		tagIDs[i] = tags[i].ID
	}

	postCounts, err := s.sqlStore.GetTagPostCounts(ctx, tagIDs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post counts")
		return
	}

	aliasMap, err := s.sqlStore.GetTagAliases(ctx, tagIDs...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag aliases")
		return
	}

	cascadeMap, err := s.sqlStore.GetTagCascades(ctx, tagIDs...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag cascades")
		return
	}

	var nextCursor *string
	if hasMore {
		encoded := obfuscateCursor(tags[len(tags)-1].Name)
		nextCursor = &encoded
	}

	items := make([]types.Tag, 0, len(tags))
	for _, tag := range tags {
		tagResp := tagFromModel(tag)
		if count, ok := postCounts[tag.ID]; ok {
			tagResp.PostCount = &count
		} else {
			zero := 0
			tagResp.PostCount = &zero
		}
		if aliases, ok := aliasMap[tag.ID]; ok {
			tagResp.Aliases = &aliases
		}
		if cascades, ok := cascadeMap[tag.ID]; ok {
			tagResp.CascadingTags = &cascades
		}
		items = append(items, tagResp)
	}

	resp := TagsResponse{
		Items:  &items,
		Cursor: nextCursor,
	}
	respond(w, http.StatusOK, resp)
}

func (s *Server) GetTag(w http.ResponseWriter, r *http.Request, name Tag) {
	ctx := r.Context()
	if name == "" {
		respondWithError(w, http.StatusBadRequest, "Tag name cannot be empty")
		return
	}
	model, err := s.sqlStore.GetTag(ctx, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Tag %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag %q", name)
		return
	}

	tagResp := tagFromModel(model)

	aliasMap, err := s.sqlStore.GetTagAliases(ctx, model.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag aliases")
		return
	}
	if aliases, ok := aliasMap[model.ID]; ok {
		tagResp.Aliases = &aliases
	}

	cascadeMap, err := s.sqlStore.GetTagCascades(ctx, model.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag cascades")
		return
	}
	if cascades, ok := cascadeMap[model.ID]; ok {
		tagResp.CascadingTags = &cascades
	}

	respond(w, http.StatusOK, tagResp)
}

func (s *Server) PutTag(w http.ResponseWriter, r *http.Request, name Tag) {
	ctx := r.Context()

	if name == "" {
		respondWithError(w, http.StatusBadRequest, "Tag name cannot be empty")
		return
	}

	var tag types.Tag
	if err := json.NewDecoder(r.Body).Decode(&tag); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !isValidName(tag.Name) {
		respondWithError(w, http.StatusBadRequest, "Tag name must begin with a unicode letter or digit")
		return
	}

	// For creates, name in body must match URL
	if tag.Name != name {
		_, err := s.sqlStore.GetTag(ctx, name)
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusBadRequest, "Tag name mismatch: got %q in body but %q in URL", tag.Name, name)
			return
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag")
			return
		}
	}

	// Resolve tag category ID if category name is provided
	logger := zerolog.Ctx(ctx).With().Str("tag", name).Logger()
	var tagCategoryID sql.Null[uuid.UUID]
	if tag.Category != nil && *tag.Category != "" {
		logger.Info().Str("category", *tag.Category).Msg("resolving tag category")
		category, err := s.sqlStore.GetTagCategory(ctx, *tag.Category)
		if err != nil {
			if errors.Is(err, store.ErrNotFound) {
				respondWithError(w, http.StatusBadRequest, "Tag category %q not found", *tag.Category)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag category")
			return
		}
		tagCategoryID = sql.Null[uuid.UUID]{V: category.ID, Valid: true}
	}

	var aliases []string
	if tag.Aliases != nil {
		aliases = *tag.Aliases
	}

	var cascadingTags []string
	if tag.CascadingTags != nil {
		cascadingTags = *tag.CascadingTags
	}

	now := time.Now().UTC()
	resultModel, isCreate, err := s.sqlStore.UpsertTag(ctx, name, store.TagInput{
		Name:          tag.Name,
		Description:   tag.Description,
		Category:      tag.Category,
		Aliases:       aliases,
		CascadingTags: cascadingTags,
		TagCategoryID: tagCategoryID,
	}, now)
	if err != nil {
		if errors.Is(err, store.ErrAliasConflict) {
			respondWithError(w, http.StatusInternalServerError, "Failed to update tag aliases")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to save tag")
		return
	}

	tagResp := tagFromModel(resultModel)

	aliasMap, err := s.sqlStore.GetTagAliases(ctx, resultModel.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag aliases")
		return
	}
	if a, ok := aliasMap[resultModel.ID]; ok {
		tagResp.Aliases = &a
	}

	cascadeMap, err := s.sqlStore.GetTagCascades(ctx, resultModel.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag cascades")
		return
	}
	if c, ok := cascadeMap[resultModel.ID]; ok {
		tagResp.CascadingTags = &c
	}

	status := http.StatusOK
	if isCreate {
		status = http.StatusCreated
		logger.Info().Msg("tag created")
	} else {
		logger.Info().Str("new_name", tag.Name).Msg("tag updated")
	}
	respond(w, status, tagResp)
}

func (s *Server) ConvertTagToAlias(w http.ResponseWriter, r *http.Request, name Tag) {
	ctx := r.Context()

	var body ConvertTagToAliasJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.Target == "" {
		respondWithError(w, http.StatusBadRequest, "Target tag name is required")
		return
	}
	if body.Target == name {
		respondWithError(w, http.StatusBadRequest, "Target must differ from source")
		return
	}

	result, err := s.sqlStore.ConvertTagToAlias(ctx, name, body.Target)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "%v", err)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to convert tag to alias")
		return
	}

	zerolog.Ctx(ctx).Info().Str("source", name).Str("target", body.Target).Msg("tag converted to alias")

	tagResp := tagFromModel(result.Tag)
	if len(result.Aliases) > 0 {
		tagResp.Aliases = &result.Aliases
	}

	respond(w, http.StatusOK, tagResp)
}

func (s *Server) DeleteTag(w http.ResponseWriter, r *http.Request, name Tag) {
	ctx := r.Context()

	if name == "" {
		respondWithError(w, http.StatusBadRequest, "Tag name cannot be empty")
		return
	}

	err := s.sqlStore.DeleteTag(ctx, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Tag %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete tag %q", name)
		return
	}
	zerolog.Ctx(ctx).Info().Str("tag", name).Msg("tag deleted")
	w.WriteHeader(http.StatusNoContent)
}
