package web

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
)

// handleTagCategories serves the tag categories listing page.
func (a *app) handleTagCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var errs []string

	cats, err := a.fetchAllTagCategories(ctx)
	if err != nil {
		errs = append(errs, fmt.Sprintf("Failed to load categories: %v", err))
	}

	// Tag counts are now provided server-side via TagCategory.TagCount.
	tagCounts := map[string]int{}
	for _, c := range cats {
		if c.TagCount != nil {
			tagCounts[c.Name] = *c.TagCount
		}
	}

	a.renderTemplate(w, r, "tag_categories", tagCategoriesData{Categories: cats, TagCounts: tagCounts, Error: strings.Join(errs, "; ")})
}

func (a *app) fetchAllTagCategories(ctx context.Context) ([]types.TagCategory, error) {
	categories := []types.TagCategory{}
	var cursor *string
	for {
		limit := 64
		resp, err := a.api.GetTagCategoriesWithResponse(ctx, &client.GetTagCategoriesParams{Limit: &limit, Cursor: cursor})
		if err != nil {
			return categories, err
		}
		if resp.JSON200 == nil {
			return categories, fmt.Errorf("API returned HTTP %d: %s", resp.StatusCode(), resp.Body)
		}
		if resp.JSON200.Items != nil {
			categories = append(categories, *resp.JSON200.Items...)
		}
		if resp.JSON200.Cursor == nil || *resp.JSON200.Cursor == "" {
			return categories, nil
		}
		cursor = resp.JSON200.Cursor
	}
}

// handleTagCategoryEdit serves the tag category edit form and handles creation, updates, and deletion.
func (a *app) handleTagCategoryEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == newResourceName

	switch r.Method {
	case http.MethodGet:
		cat := types.TagCategory{Color: "#888888"}
		var editErr string
		if !isNew {
			resp, err := a.api.GetTagCategoryWithResponse(ctx, name)
			if err != nil {
				editErr = fmt.Sprintf("Failed to load category: %v", err)
			} else if resp.JSON200 != nil {
				cat = *resp.JSON200
			} else {
				editErr = fmt.Sprintf("Failed to load category: %s", resp.Body)
			}
		}
		a.renderTemplate(w, r, "tag_category_edit", tagCategoryEditData{
			Category:    cat,
			CurrentName: name,
			IsNew:       isNew,
			Error:       editErr,
		})

	case http.MethodPost:
		newName := r.FormValue("name")
		description := r.FormValue("description")
		color := r.FormValue("color")

		cat := types.TagCategory{
			Name:        newName,
			Description: description,
			Color:       color,
		}

		urlName := name
		if isNew {
			urlName = newName
		}
		resp, err := a.api.PutTagCategoryWithResponse(ctx, urlName, client.NewTagCategoryUpdateRequest(cat))
		if err != nil || resp.StatusCode() >= 400 {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Failed to save category: %v", err)
			} else {
				errMsg = fmt.Sprintf("Failed to save category: %s", resp.Body)
			}
			a.renderTemplate(w, r, "tag_category_edit", tagCategoryEditData{
				Category:    cat,
				CurrentName: name,
				IsNew:       isNew,
				Error:       errMsg,
			})
			return
		}
		http.Redirect(w, r, "/tag-categories", http.StatusSeeOther)

	case http.MethodDelete:
		resp, err := a.api.DeleteTagCategoryWithResponse(ctx, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete category: %v", err), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode() >= 400 {
			http.Error(w, fmt.Sprintf("Failed to delete category: %s", resp.Body), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tag-categories", http.StatusSeeOther)
	}
}
