package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/spf13/cobra"
)

type editableTag struct {
	Name        string  `yaml:"name"`
	Category    *string `yaml:"category,omitempty"`
	Description string  `yaml:"description"`
}

func init() {
	getTagCmd := &cobra.Command{
		Use:   "tag [name]",
		Short: "Get a tag by name, or list all tags",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getTag(args[0])
			}
			return listTags()
		},
	}

	var tagDescription string
	var tagCategory string

	createTagCmd := &cobra.Command{
		Use:   "tag <name>",
		Short: "Create a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tag := types.Tag{
				Name:        args[0],
				Description: tagDescription,
			}
			if cmd.Flags().Changed("category") {
				tag.Category = &tagCategory
			}
			return createTag(args[0], tag)
		},
	}
	createTagCmd.Flags().StringVar(&tagDescription, "description", "", "Tag description")
	createTagCmd.Flags().StringVar(&tagCategory, "category", "", "Tag category name")

	editTagCmd := &cobra.Command{
		Use:   "tag <name>",
		Short: "Edit a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return editTag(args[0])
		},
	}

	deleteTagCmd := &cobra.Command{
		Use:   "tag <name>",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteTag(args[0])
		},
	}

	getCmd.AddCommand(getTagCmd)
	createCmd.AddCommand(createTagCmd)
	editCmd.AddCommand(editTagCmd)
	deleteCmd.AddCommand(deleteTagCmd)
}

func getTag(name string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/tags/%s", cfg.APIURL, url.PathEscape(name)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	tag, err := decodeJSON[types.Tag](resp)
	if err != nil {
		return err
	}
	return printResource(tag, func() [][2]string {
		cat := "(none)"
		if tag.Category != nil {
			cat = *tag.Category
		}
		return [][2]string{
			{"Name", tag.Name},
			{"Category", cat},
			{"Description", tag.Description},
			{"Created", tag.CreatedAt.Format(time.RFC3339)},
			{"Updated", tag.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func listTags() error {
	tags, err := fetchAll[types.Tag](cfg, cfg.APIURL+"/api/v1/tags", url.Values{})
	if err != nil {
		return err
	}
	return printList(tags, []string{"NAME", "CATEGORY", "DESCRIPTION"}, func() [][]string {
		rows := make([][]string, 0, len(tags))
		for _, t := range tags {
			cat := "(none)"
			if t.Category != nil {
				cat = *t.Category
			}
			rows = append(rows, []string{t.Name, cat, t.Description})
		}
		return rows
	})
}

func createTag(name string, tag types.Tag) error {
	body, err := json.Marshal(tag)
	if err != nil {
		return fmt.Errorf("marshal tag: %w", err)
	}
	resp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/tags/%s", cfg.APIURL, url.PathEscape(name)), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	created, err := decodeJSON[types.Tag](resp)
	if err != nil {
		return err
	}
	return printResource(created, func() [][2]string {
		cat := "(none)"
		if created.Category != nil {
			cat = *created.Category
		}
		return [][2]string{
			{"Name", created.Name},
			{"Category", cat},
			{"Description", created.Description},
			{"Created", created.CreatedAt.Format(time.RFC3339)},
			{"Updated", created.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func editTag(name string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/tags/%s", cfg.APIURL, url.PathEscape(name)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	tag, err := decodeJSON[types.Tag](resp)
	if err != nil {
		return err
	}

	editable := editableTag{
		Name:        tag.Name,
		Category:    tag.Category,
		Description: tag.Description,
	}

	var edited editableTag
	changed, err := openEditor(editable, &edited)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Println("No changes.")
		return nil
	}

	updated := types.Tag{
		Name:        edited.Name,
		Category:    edited.Category,
		Description: edited.Description,
	}
	body, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf("marshal tag: %w", err)
	}
	putResp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/tags/%s", cfg.APIURL, url.PathEscape(name)), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = putResp.Body.Close() }()
	if err := checkStatus(putResp); err != nil {
		return err
	}
	result, err := decodeJSON[types.Tag](putResp)
	if err != nil {
		return err
	}
	return printResource(result, func() [][2]string {
		cat := "(none)"
		if result.Category != nil {
			cat = *result.Category
		}
		return [][2]string{
			{"Name", result.Name},
			{"Category", cat},
			{"Description", result.Description},
			{"Created", result.CreatedAt.Format(time.RFC3339)},
			{"Updated", result.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func deleteTag(name string) error {
	resp, err := doRequest(cfg, http.MethodDelete, fmt.Sprintf("%s/api/v1/tags/%s", cfg.APIURL, url.PathEscape(name)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	fmt.Printf("tag/%s deleted\n", name)
	return nil
}
