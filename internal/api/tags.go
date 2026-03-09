package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/dharmab/hyperboard/internal/db/models"
	"github.com/dharmab/hyperboard/internal/types"
	"github.com/gofrs/uuid/v5"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/dialect"
	"github.com/stephenafamo/bob/dialect/psql/dm"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// isValidTagName reports whether name begins with a unicode letter or digit.
func isValidTagName(name string) bool {
	for _, r := range name {
		return unicode.IsLetter(r) || unicode.IsDigit(r)
	}
	return false
}

func tagFromModel(model *models.Tag) types.Tag {
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

	return tag
}

// getTagAliases returns a map from tag ID to its list of aliases.
func (s *Server) getTagAliases(ctx context.Context, tagIDs ...uuid.UUID) (map[uuid.UUID][]string, error) {
	if len(tagIDs) == 0 {
		return map[uuid.UUID][]string{}, nil
	}

	args := make([]any, len(tagIDs))
	var placeholders strings.Builder
	for i, id := range tagIDs {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT tag_id, alias_name FROM tag_aliases WHERE tag_id IN ("+placeholders.String()+") ORDER BY alias_name",
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	result := make(map[uuid.UUID][]string)
	for rows.Next() {
		var tagID uuid.UUID
		var alias string
		if err := rows.Scan(&tagID, &alias); err != nil {
			return nil, err
		}
		result[tagID] = append(result[tagID], alias)
	}
	return result, rows.Err()
}

// setTagAliases replaces all aliases for a tag with the given list.
// Returns an error if any alias conflicts with an existing tag name.
func (s *Server) setTagAliases(ctx context.Context, tagID uuid.UUID, aliases []string) error {
	// Check that no alias conflicts with an existing tag name
	for _, alias := range aliases {
		if alias == "" {
			continue
		}
		var count int
		err := s.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM tags WHERE name = $1", alias,
		).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("alias %q conflicts with an existing tag name", alias)
		}
	}

	_, err := s.db.ExecContext(ctx, "DELETE FROM tag_aliases WHERE tag_id = $1", tagID)
	if err != nil {
		return err
	}
	for _, alias := range aliases {
		if alias == "" {
			continue
		}
		_, err := s.db.ExecContext(ctx,
			"INSERT INTO tag_aliases (tag_id, alias_name) VALUES ($1, $2)",
			tagID, alias,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// resolveAlias looks up an alias and returns the canonical tag name.
// If the name is not an alias, it is returned as-is.
func (s *Server) resolveAlias(ctx context.Context, name string) (string, error) {
	var canonical string
	err := s.db.QueryRowContext(ctx,
		"SELECT t.name FROM tags t JOIN tag_aliases ta ON t.id = ta.tag_id WHERE ta.alias_name = $1",
		name,
	).Scan(&canonical)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return name, nil
		}
		return "", err
	}
	return canonical, nil
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

	// Collect tag IDs for the current page only
	pageSize := min(len(tags), limit)
	tagIDs := make([]uuid.UUID, pageSize)
	for i := range pageSize {
		tagIDs[i] = tags[i].ID
	}

	// Query post counts only for the tags on this page
	postCounts, err := s.getTagPostCounts(ctx, tagIDs)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve post counts")
		return
	}

	// Query aliases only for the tags on this page
	aliasMap, err := s.getTagAliases(ctx, tagIDs...)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve tag aliases")
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

	if model.TagCategoryID.Valid {
		if err := model.LoadTagCategory(ctx, s.db); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to load tag category")
			return
		}
	}

	tagResp := tagFromModel(model)

	aliasMap, err := s.getTagAliases(ctx, model.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag aliases")
		return
	}
	if aliases, ok := aliasMap[model.ID]; ok {
		tagResp.Aliases = &aliases
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

	if !isValidTagName(tag.Name) {
		respondWithError(w, http.StatusBadRequest, "Tag name must begin with a unicode letter or digit")
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

	now := new(time.Now().UTC())
	var resultModel *models.Tag
	if existing != nil {
		// Update (supports rename)
		err = existing.Update(ctx, s.db, &models.TagSetter{
			Name:          &tag.Name,
			Description:   &tag.Description,
			TagCategoryID: &tagCategoryID,
			UpdatedAt:     now,
		})
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to update tag")
			return
		}
		existing.Name = tag.Name
		existing.Description = tag.Description
		existing.TagCategoryID = tagCategoryID
		resultModel = existing
	} else {
		if tag.Name != name {
			respondWithError(w, http.StatusBadRequest, "Tag name mismatch: got %q in body but %q in URL", tag.Name, name)
			return
		}
		resultModel, err = models.Tags.Insert(
			&models.TagSetter{
				Name:          &tag.Name,
				Description:   &tag.Description,
				TagCategoryID: &tagCategoryID,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		).One(ctx, s.db)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create tag")
			return
		}
	}

	// Update aliases
	var aliases []string
	if tag.Aliases != nil {
		aliases = *tag.Aliases
	}
	if err := s.setTagAliases(ctx, resultModel.ID, aliases); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update tag aliases")
		return
	}

	if resultModel.TagCategoryID.Valid {
		if err := resultModel.LoadTagCategory(ctx, s.db); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to load tag category")
			return
		}
	}

	tagResp := tagFromModel(resultModel)

	aliasMap, err := s.getTagAliases(ctx, resultModel.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag aliases")
		return
	}
	if a, ok := aliasMap[resultModel.ID]; ok {
		tagResp.Aliases = &a
	}

	status := http.StatusOK
	if existing == nil {
		status = http.StatusCreated
	}
	respond(w, status, tagResp)
}

// getTagPostCounts returns post counts for the given tag IDs.
func (s *Server) getTagPostCounts(ctx context.Context, tagIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	if len(tagIDs) == 0 {
		return map[uuid.UUID]int{}, nil
	}

	args := make([]any, len(tagIDs))
	var placeholders strings.Builder
	for i, id := range tagIDs {
		if i > 0 {
			placeholders.WriteString(", ")
		}
		placeholders.WriteString("$" + strconv.Itoa(i+1))
		args[i] = id
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT tag_id, COUNT(*) FROM posts_tags WHERE tag_id IN ("+placeholders.String()+") GROUP BY tag_id",
		args...,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

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

func (s *Server) ConvertTagToAlias(w http.ResponseWriter, r *http.Request, name Tag) {
	ctx := r.Context()

	var body ConvertTagToAliasJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if body.Target == "" || body.Target == name {
		respondWithError(w, http.StatusBadRequest, "Invalid target tag")
		return
	}

	tx, err := s.db.DB.BeginTx(ctx, nil)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to begin transaction")
		return
	}
	defer func() { _ = tx.Rollback() }()

	// Look up source tag
	var sourceID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", name).Scan(&sourceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Source tag %q not found", name)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to look up source tag")
		return
	}

	// Look up target tag
	var targetID uuid.UUID
	err = tx.QueryRowContext(ctx, "SELECT id FROM tags WHERE name = $1", body.Target).Scan(&targetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusNotFound, "Target tag %q not found", body.Target)
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to look up target tag")
		return
	}

	// Re-tag posts: move source associations to target where target doesn't already exist
	_, err = tx.ExecContext(ctx,
		`UPDATE posts_tags SET tag_id = $1
		 WHERE tag_id = $2
		   AND NOT EXISTS (SELECT 1 FROM posts_tags pt2 WHERE pt2.post_id = posts_tags.post_id AND pt2.tag_id = $1)`,
		targetID, sourceID,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to re-tag posts")
		return
	}

	// Delete remaining source associations (posts that already had the target tag)
	_, err = tx.ExecContext(ctx, "DELETE FROM posts_tags WHERE tag_id = $1", sourceID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to clean up source tag associations")
		return
	}

	// Collect source aliases before deleting
	aliasRows, err := tx.QueryContext(ctx, "SELECT alias_name FROM tag_aliases WHERE tag_id = $1", sourceID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to collect source aliases")
		return
	}
	defer func() { _ = aliasRows.Close() }()
	var sourceAliases []string
	for aliasRows.Next() {
		var alias string
		if err := aliasRows.Scan(&alias); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to read source aliases")
			return
		}
		sourceAliases = append(sourceAliases, alias)
	}
	if err := aliasRows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to iterate source aliases")
		return
	}

	// Delete source tag (cascades aliases)
	_, err = tx.ExecContext(ctx, "DELETE FROM tags WHERE id = $1", sourceID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to delete source tag")
		return
	}

	// Add source name + source aliases as aliases of target
	allNewAliases := append([]string{name}, sourceAliases...)
	for _, alias := range allNewAliases {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO tag_aliases (tag_id, alias_name) VALUES ($1, $2) ON CONFLICT DO NOTHING",
			targetID, alias,
		)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to add alias %q", alias)
			return
		}
	}

	// Read the target tag and its aliases within the transaction so the
	// response is consistent even if a concurrent request modifies it.
	var tagName, tagDesc string
	var tagCatID sql.Null[uuid.UUID]
	var tagCreatedAt, tagUpdatedAt time.Time
	err = tx.QueryRowContext(ctx,
		"SELECT name, description, tag_category_id, created_at, updated_at FROM tags WHERE id = $1",
		targetID,
	).Scan(&tagName, &tagDesc, &tagCatID, &tagCreatedAt, &tagUpdatedAt)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve updated target tag")
		return
	}

	var catName *string
	if tagCatID.Valid {
		var cn string
		err = tx.QueryRowContext(ctx, "SELECT name FROM tag_categories WHERE id = $1", tagCatID.V).Scan(&cn)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			respondWithError(w, http.StatusInternalServerError, "Failed to load tag category")
			return
		}
		if err == nil {
			catName = &cn
		}
	}

	txAliasRows, err := tx.QueryContext(ctx, "SELECT alias_name FROM tag_aliases WHERE tag_id = $1 ORDER BY alias_name", targetID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to load tag aliases")
		return
	}
	defer func() { _ = txAliasRows.Close() }()
	var targetAliases []string
	for txAliasRows.Next() {
		var a string
		if err := txAliasRows.Scan(&a); err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to read tag aliases")
			return
		}
		targetAliases = append(targetAliases, a)
	}
	if err := txAliasRows.Err(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to iterate tag aliases")
		return
	}

	if err := tx.Commit(); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to commit transaction")
		return
	}

	tagResp := types.Tag{
		Name:        tagName,
		Description: tagDesc,
		Category:    catName,
		CreatedAt:   tagCreatedAt,
		UpdatedAt:   tagUpdatedAt,
	}
	if len(targetAliases) > 0 {
		tagResp.Aliases = &targetAliases
	}

	respond(w, http.StatusOK, tagResp)
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
