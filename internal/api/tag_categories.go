package api

import (
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

func tagCategoryFromModel(model *models.TagCategory) types.TagCategory {
	return types.TagCategory{
		Name:        model.Name,
		Description: model.Description,
		Color:       model.Color,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func (s *Server) GetTagCategories(w http.ResponseWriter, r *http.Request, params GetTagCategoriesParams) {
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

	categories, hasMore, err := s.sqlStore.ListTagCategories(ctx, decodedCursor, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag categories")
		return
	}

	// Collect IDs for the current page
	catIDs := make([]uuid.UUID, len(categories))
	for i := range categories {
		catIDs[i] = categories[i].ID
	}

	tagCounts, err := s.sqlStore.GetTagCountsByCategory(ctx, catIDs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag counts")
		return
	}

	var nextCursor *string
	if hasMore {
		encoded := obfuscateCursor(categories[len(categories)-1].Name)
		nextCursor = &encoded
	}

	items := make([]types.TagCategory, 0, len(categories))
	for _, category := range categories {
		cat := tagCategoryFromModel(category)
		if count, ok := tagCounts[category.ID]; ok {
			cat.TagCount = &count
		} else {
			zero := 0
			cat.TagCount = &zero
		}
		items = append(items, cat)
	}
	resp := TagCategoriesResponse{
		Items:  &items,
		Cursor: nextCursor,
	}
	respond(w, http.StatusOK, resp)
}

func (s *Server) GetTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
	ctx := r.Context()
	model, err := s.sqlStore.GetTagCategory(ctx, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Tag category %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag category %q", name)
		return
	}
	respond(w, http.StatusOK, tagCategoryFromModel(model))
}

func (s *Server) PutTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
	ctx := r.Context()
	var req types.TagCategory
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if !isValidName(req.Name) {
		respondWithError(w, http.StatusBadRequest, "Tag category name must begin with a unicode letter or digit")
		return
	}

	// For creates, name in body must match URL
	if req.Name != name {
		// Check if this is an update (existing category) - if so, rename is allowed
		_, err := s.sqlStore.GetTagCategory(ctx, name)
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusBadRequest, "Tag category name mismatch: got %q in body but %q in URL", req.Name, name)
			return
		}
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag category")
			return
		}
	}

	if req.Color != "" && !isValidHexColor(req.Color) {
		respondWithError(w, http.StatusBadRequest, "Color must be a valid hex color (e.g. #ff0000)")
		return
	}
	if req.Color == "" {
		req.Color = "#888888"
	}

	logger := zerolog.Ctx(ctx).With().Str("category", name).Logger()
	now := time.Now().UTC()
	model, isCreate, err := s.sqlStore.UpsertTagCategory(ctx, name, store.TagCategoryInput{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}, now)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to save tag category")
		return
	}

	if isCreate {
		logger.Info().Msg("tag category created")
		respond(w, http.StatusCreated, tagCategoryFromModel(model))
	} else {
		logger.Info().Str("new_name", req.Name).Msg("tag category updated")
		respond(w, http.StatusOK, tagCategoryFromModel(model))
	}
}

func (s *Server) DeleteTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
	ctx := r.Context()
	err := s.sqlStore.DeleteTagCategory(ctx, name)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Tag category %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete tag category %q", name)
		return
	}
	zerolog.Ctx(ctx).Info().Str("category", name).Msg("tag category deleted")
	w.WriteHeader(http.StatusNoContent)
}
