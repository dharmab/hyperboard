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
	"github.com/stephenafamo/bob/dialect/psql/im"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func tagCategoryFromModel(model *models.TagCategory) types.TagCategory {
	return types.TagCategory{
		Name:      model.Name,
		CreatedAt: model.CreatedAt.V,
		UpdatedAt: model.UpdatedAt.V,
	}
}

func (s *Server) GetTagCategories(w http.ResponseWriter, r *http.Request, params GetTagCategoriesParams) {
	ctx := r.Context()

	// Ordering
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.OrderBy(models.TagCategories.Name()).Asc(),
	}

	// Cursor
	if params.Cursor != nil && *params.Cursor != "" {
		decodedName, err := deobfuscateCursor(*params.Cursor)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		mods = append(mods, sm.Where(models.TagCategories.Name().GT(psql.Arg(decodedName))))
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
		sm.Where(models.TagCategories.Name().EQ(psql.Arg(name))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Tag category %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag category %q", name)
		return
	}
	respond(w, http.StatusOK, model)
}

func (s *Server) PutTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
	ctx := r.Context()
	var tagCategory types.TagCategory
	if err := json.NewDecoder(r.Body).Decode(&tagCategory); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if tagCategory.Name != name {
		respondWithError(w, http.StatusBadRequest, "Tag category name mismatch: got %q in body but %q in URL", tagCategory.Name, name)
		return
	}
	upsertedModel, err := models.TagCategories.Insert(
		&models.TagCategorySetter{
			Name:      &tagCategory.Name,
			CreatedAt: now(),
			UpdatedAt: now(),
		},
		im.OnConflict(models.TagCategoryColumns.Name).DoUpdate(
			im.SetExcluded(models.TagCategoryColumns.CreatedAt.String()),
		),
	).One(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to store tag category")
		return
	}
	respond(w, http.StatusOK, tagCategoryFromModel(upsertedModel))
}

func (s *Server) DeleteTagCategory(w http.ResponseWriter, r *http.Request, name TagCategory) {
	ctx := r.Context()
	_, err := models.TagCategories.Delete(
		dm.Where(models.TagCategories.Name().EQ(psql.Arg(name))),
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
