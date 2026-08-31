package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
)

// handleTags serves the tags listing page.
func (a *app) handleTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Fetch all tags (paginate through all pages)
	var errs []string
	allTags := []types.Tag{}
	var cursor *string
	for {
		limit := 1000
		params := &client.GetTagsParams{Limit: &limit, Cursor: cursor}
		resp, err := a.api.GetTagsWithResponse(ctx, params)
		if err != nil {
			errs = append(errs, fmt.Sprintf("Failed to load tags: %v", err))
			break
		}
		if resp.StatusCode() >= 400 {
			errs = append(errs, fmt.Sprintf("Failed to load tags: %s", resp.Body))
			break
		}
		if resp.JSON200 != nil && resp.JSON200.Items != nil {
			allTags = append(allTags, *resp.JSON200.Items...)
		}
		if resp.JSON200 == nil || resp.JSON200.Cursor == nil || *resp.JSON200.Cursor == "" {
			break
		}
		cursor = resp.JSON200.Cursor
	}

	// Fetch categories for color map
	categories, err := a.fetchAllTagCategories(ctx)
	if err != nil {
		errs = append(errs, fmt.Sprintf("Failed to load categories: %v", err))
	}
	colorMap := map[string]string{}
	for _, category := range categories {
		colorMap[category.Name] = category.Color
	}

	a.renderTemplate(w, r, "tags", tagsData{Tags: allTags, CategoryColors: colorMap, Error: strings.Join(errs, "; ")})
}

// handleTagEdit serves the tag edit form and handles tag creation, updates, and deletion.
func (a *app) handleTagEdit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	name := r.PathValue("name")
	isNew := name == newResourceName

	// Fetch categories for dropdown
	cats, err := a.fetchAllTagCategories(ctx)
	var catErr string
	if err != nil {
		catErr = fmt.Sprintf("Failed to load categories: %v", err)
	}

	switch r.Method {
	case http.MethodGet:
		tag := types.Tag{}
		var errs []string
		if catErr != "" {
			errs = append(errs, catErr)
		}
		if !isNew {
			resp, err := a.api.GetTagWithResponse(ctx, name)
			if err != nil {
				errs = append(errs, fmt.Sprintf("Failed to load tag: %v", err))
			} else if resp.JSON200 != nil {
				tag = *resp.JSON200
			} else {
				errs = append(errs, fmt.Sprintf("Failed to load tag: %s", resp.Body))
			}
		}
		var aliases []string
		if tag.Aliases != nil {
			aliases = *tag.Aliases
		}
		var cascadingTags []string
		if tag.CascadingTags != nil {
			cascadingTags = *tag.CascadingTags
		}
		a.renderTemplate(w, r, "tag_edit", tagEditData{
			Tag:           tag,
			Aliases:       aliases,
			CascadingTags: cascadingTags,
			Categories:    cats,
			CurrentName:   name,
			IsNew:         isNew,
			Error:         strings.Join(errs, "; "),
		})

	case http.MethodPost:
		newName := r.FormValue("name")
		description := r.FormValue("description")
		category := r.FormValue("category")
		aliasesRaw := r.FormValue("aliases")
		cascadingTagsRaw := r.FormValue("cascading_tags")

		var aliases []string
		for a := range strings.SplitSeq(aliasesRaw, ",") {
			a = strings.TrimSpace(a)
			if a != "" {
				aliases = append(aliases, a)
			}
		}

		var cascadingTags []string
		for ct := range strings.SplitSeq(cascadingTagsRaw, ",") {
			ct = strings.TrimSpace(ct)
			if ct != "" {
				cascadingTags = append(cascadingTags, ct)
			}
		}

		tag := types.Tag{
			Name:          newName,
			Description:   description,
			Aliases:       &aliases,
			CascadingTags: &cascadingTags,
		}
		if category != "" {
			tag.Category = &category
		}

		urlName := name
		if isNew {
			urlName = newName
		}
		resp, err := a.api.PutTagWithResponse(ctx, urlName, client.NewTagUpdateRequest(tag))
		if err != nil || resp.StatusCode() >= 400 {
			var errMsg string
			if err != nil {
				errMsg = fmt.Sprintf("Failed to save tag: %v", err)
			} else {
				errMsg = fmt.Sprintf("Failed to save tag: %s", resp.Body)
			}
			a.renderTemplate(w, r, "tag_edit", tagEditData{
				Tag:           tag,
				Aliases:       aliases,
				CascadingTags: cascadingTags,
				Categories:    cats,
				CurrentName:   name,
				IsNew:         isNew,
				Error:         errMsg,
			})
			return
		}
		http.Redirect(w, r, "/tags", http.StatusSeeOther)

	case http.MethodDelete:
		resp, err := a.api.DeleteTagWithResponse(ctx, name)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to delete tag: %v", err), http.StatusInternalServerError)
			return
		}
		if resp.StatusCode() >= 400 {
			http.Error(w, fmt.Sprintf("Failed to delete tag: %s", resp.Body), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/tags", http.StatusSeeOther)
	}
}

func (a *app) handleTagConvertToAlias(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sourceName := r.PathValue("name")
	targetName := r.FormValue("target")

	if targetName == "" || sourceName == targetName {
		http.Error(w, "Invalid target tag", http.StatusBadRequest)
		return
	}

	resp, err := a.api.ConvertTagToAliasWithResponse(ctx, sourceName, client.ConvertTagToAliasJSONRequestBody{
		Target: targetName,
	})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to convert tag to alias: %v", err), http.StatusInternalServerError)
		return
	}
	if resp.StatusCode() >= 400 {
		http.Error(w, fmt.Sprintf("Failed to convert tag to alias: %s", resp.Body), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/tags/"+escapePathSegment(targetName), http.StatusSeeOther)
}
