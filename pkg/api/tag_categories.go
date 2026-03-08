package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
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

	// Ordering
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.OrderBy(models.TagCategoryColumns.Name).Asc(),
	}

	// Cursor
	if params.Cursor != nil && *params.Cursor != "" {
		decodedName, err := deobfuscateCursor(*params.Cursor)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		mods = append(mods, sm.Where(models.TagCategoryColumns.Name.GT(psql.Arg(decodedName))))
	}

	// Limit
	limit := parseLimit(params.Limit)
	mods = append(mods, sm.Limit(int64(limit+1)))

	// Query
	categories, err := models.TagCategories.Query(mods...).All(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag categories")
		return
	}

	// Check if there's content after the limit
	hasMore, nextCursor := paginate(len(categories), limit, func() string {
		return categories[limit-1].Name
	})
	if hasMore {
		categories = categories[:limit]
	}

	// Response
	items := make([]types.TagCategory, 0, len(categories))
	for _, category := range categories {
		items = append(items, tagCategoryFromModel(category))
	}
	resp := TagCategoriesResponse{
		Items:  &items,
		Cursor: nextCursor,
	}
	respond(w, http.StatusOK, resp)
}

func (s *Server) GetTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
	ctx := r.Context()
	model, err := models.TagCategories.Query(
		sm.Where(models.TagCategoryColumns.Name.EQ(psql.Arg(name))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
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

	existing, err := models.TagCategories.Query(
		sm.Where(models.TagCategoryColumns.Name.EQ(psql.Arg(name))),
	).One(ctx, s.db)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag category")
		return
	}

	if existing != nil {
		// Update (supports rename)
		err = existing.Update(ctx, s.db, &models.TagCategorySetter{
			Name:        &req.Name,
			Description: &req.Description,
			Color:       &req.Color,
			UpdatedAt:   now(),
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update tag category")
			return
		}
		existing.Name = req.Name
		existing.Description = req.Description
		existing.Color = req.Color
		respond(w, http.StatusOK, tagCategoryFromModel(existing))
	} else {
		if req.Name != name {
			respondWithError(w, http.StatusBadRequest, "Tag category name mismatch: got %q in body but %q in URL", req.Name, name)
			return
		}
		inserted, err := models.TagCategories.Insert(
			&models.TagCategorySetter{
				Name:        &req.Name,
				Description: &req.Description,
				Color:       &req.Color,
				CreatedAt:   now(),
				UpdatedAt:   now(),
			},
		).One(ctx, s.db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create tag category")
			return
		}
		respond(w, http.StatusCreated, tagCategoryFromModel(inserted))
	}
}

func (s *Server) DeleteTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
	ctx := r.Context()
	_, err := models.TagCategories.Delete(
		dm.Where(models.TagCategoryColumns.Name.EQ(psql.Arg(name))),
	).Exec(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Tag category %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete tag category %q", name)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
