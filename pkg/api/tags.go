package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

func tagFromModel(model *models.Tag) (types.Tag, error) {
	tag := types.Tag{
		Name:        model.Name,
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	// Load the tag category if present
	if model.TagCategoryID.Valid {
		if model.R.TagCategory != nil {
			tag.Category = &model.R.TagCategory.Name
		}
	}

	return tag, nil
}

func (s *Server) GetTags(w http.ResponseWriter, r *http.Request, params GetTagsParams) {
	ctx := r.Context()

	// Ordering
	mods := []bob.Mod[*dialect.SelectQuery]{
		sm.OrderBy(models.TagColumns.Name).Asc(),
	}

	// Cursor
	if params.Cursor != nil && *params.Cursor != "" {
		decodedName, err := deobfuscateCursor(*params.Cursor)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid cursor")
			return
		}
		mods = append(mods, sm.Where(models.TagColumns.Name.GT(psql.Arg(decodedName))))
	}

	// Limit
	limit := parseLimit(params.Limit)
	mods = append(mods, sm.Limit(int64(limit+1)))

	tags, err := models.Tags.Query(mods...).All(ctx, s.db)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tags")
		return
	}

	// Load tag category relationships
	if err := tags.LoadTagCategory(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag categories")
		return
	}

	// Query post counts per tag
	postCounts, err := s.getTagPostCounts(ctx)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post counts")
		return
	}

	// Check if there's content after the limit
	hasMore, nextCursor := paginate(len(tags), limit, func() string {
		return tags[limit-1].Name
	})
	if hasMore {
		tags = tags[:limit]
	}

	items := make([]types.Tag, 0, len(tags))
	for _, tag := range tags {
		tagResp, err := tagFromModel(tag)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to convert tag")
			return
		}
		if count, ok := postCounts[tag.ID]; ok {
			tagResp.PostCount = &count
		} else {
			zero := 0
			tagResp.PostCount = &zero
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
	model, err := models.Tags.Query(
		sm.Where(models.TagColumns.Name.EQ(psql.Arg(name))),
	).One(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Tag %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag %q", name)
		return
	}

	if err := model.LoadTagCategory(ctx, s.db); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag category")
		return
	}

	tagResp, err := tagFromModel(model)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to convert tag")
		return
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

	// Resolve tag category ID if category name is provided
	var tagCategoryID sql.Null[uuid.UUID]
	if tag.Category != nil && *tag.Category != "" {
		category, err := models.TagCategories.Query(
			sm.Where(models.TagCategoryColumns.Name.EQ(psql.Arg(*tag.Category))),
		).One(ctx, s.db)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				respondWithError(w, http.StatusBadRequest, "Tag category %q not found", *tag.Category)
				return
			}
			respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag category")
			return
		}
		tagCategoryID = sql.Null[uuid.UUID]{V: category.ID, Valid: true}
	}

	existing, err := models.Tags.Query(
		sm.Where(models.TagColumns.Name.EQ(psql.Arg(name))),
	).One(ctx, s.db)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag")
		return
	}

	var resultModel *models.Tag
	if existing != nil {
		// Update (supports rename)
		err = existing.Update(ctx, s.db, &models.TagSetter{
			Name:          &tag.Name,
			Description:   &tag.Description,
			TagCategoryID: &tagCategoryID,
			UpdatedAt:     now(),
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update tag")
			return
		}
		existing.Name = tag.Name
		existing.Description = tag.Description
		existing.TagCategoryID = tagCategoryID
		resultModel = existing
		if err := resultModel.LoadTagCategory(ctx, s.db); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to load tag category")
			return
		}
		tagResp, err := tagFromModel(resultModel)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to convert tag")
			return
		}
		respond(w, http.StatusOK, tagResp)
	} else {
		if tag.Name != name {
			respondWithError(w, http.StatusBadRequest, "Tag name mismatch: got %q in body but %q in URL", tag.Name, name)
			return
		}
		inserted, err := models.Tags.Insert(
			&models.TagSetter{
				Name:          &tag.Name,
				Description:   &tag.Description,
				TagCategoryID: &tagCategoryID,
				CreatedAt:     now(),
				UpdatedAt:     now(),
			},
		).One(ctx, s.db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create tag")
			return
		}
		if err := inserted.LoadTagCategory(ctx, s.db); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to load tag category")
			return
		}
		tagResp, err := tagFromModel(inserted)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to convert tag")
			return
		}
		respond(w, http.StatusCreated, tagResp)
	}
}

func (s *Server) getTagPostCounts(ctx context.Context) (map[uuid.UUID]int, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT tag_id, COUNT(*) FROM posts_tags GROUP BY tag_id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := make(map[uuid.UUID]int)
	for rows.Next() {
		var tagID uuid.UUID
		var count int
		if err := rows.Scan(&tagID, &count); err != nil {
			return nil, err
		}
		counts[tagID] = count
	}
	return counts, rows.Err()
}

func (s *Server) DeleteTag(w http.ResponseWriter, r *http.Request, name Tag) {
	ctx := r.Context()

	if name == "" {
		respondWithError(w, http.StatusBadRequest, "Tag name cannot be empty")
		return
	}

	_, err := models.Tags.Delete(
		dm.Where(models.TagColumns.Name.EQ(psql.Arg(name))),
	).Exec(ctx, s.db)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Tag %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete tag %q", name)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
