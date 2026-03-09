package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/dharmab/hyperboard/internal/types"
	"github.com/spf13/cobra"
)

type editableTagCategory struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Color       string `yaml:"color"`
}

func init() {
	getTagCategoryCmd := &cobra.Command{
		Use:   "tagcategory [name]",
		Short: "Get a tag category by name, or list all tag categories",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getTagCategory(args[0])
			}
			return listTagCategories()
		},
	}

	var tcDescription string
	var tcColor string

	createTagCategoryCmd := &cobra.Command{
		Use:   "tagcategory <name>",
		Short: "Create a tag category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tc := types.TagCategory{
				Name:        args[0],
				Description: tcDescription,
				Color:       tcColor,
			}
			return createTagCategory(args[0], tc)
		},
	}
	createTagCategoryCmd.Flags().StringVar(&tcDescription, "description", "", "Tag category description")
	createTagCategoryCmd.Flags().StringVar(&tcColor, "color", "#888888", "Tag category color (hex)")

	editTagCategoryCmd := &cobra.Command{
		Use:   "tagcategory <name>",
		Short: "Edit a tag category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return editTagCategory(args[0])
		},
	}

	deleteTagCategoryCmd := &cobra.Command{
		Use:   "tagcategory <name>",
		Short: "Delete a tag category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteTagCategory(args[0])
		},
	}

	getCmd.AddCommand(getTagCategoryCmd)
	createCmd.AddCommand(createTagCategoryCmd)
	editCmd.AddCommand(editTagCategoryCmd)
	deleteCmd.AddCommand(deleteTagCategoryCmd)
}

func getTagCategory(name string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/tagCategories/%s", cfg.APIURL, url.PathEscape(name)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	tc, err := decodeJSON[types.TagCategory](resp)
	if err != nil {
		return err
	}
	return printResource(tc, func() [][2]string {
		return [][2]string{
			{"Name", tc.Name},
			{"Description", tc.Description},
			{"Color", tc.Color},
			{"Created", tc.CreatedAt.Format(time.RFC3339)},
			{"Updated", tc.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func listTagCategories() error {
	tcs, err := fetchAll[types.TagCategory](cfg, cfg.APIURL+"/api/v1/tagCategories", url.Values{})
	if err != nil {
		return err
	}
	return printList(tcs, []string{"NAME", "DESCRIPTION", "COLOR"}, func() [][]string {
		rows := make([][]string, 0, len(tcs))
		for _, tc := range tcs {
			rows = append(rows, []string{tc.Name, tc.Description, tc.Color})
		}
		return rows
	})
}

func createTagCategory(name string, tc types.TagCategory) error {
	body, err := json.Marshal(tc)
	if err != nil {
		return fmt.Errorf("marshal tag category: %w", err)
	}
	resp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/tagCategories/%s", cfg.APIURL, url.PathEscape(name)), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	created, err := decodeJSON[types.TagCategory](resp)
	if err != nil {
		return err
	}
	return printResource(created, func() [][2]string {
		return [][2]string{
			{"Name", created.Name},
			{"Description", created.Description},
			{"Color", created.Color},
			{"Created", created.CreatedAt.Format(time.RFC3339)},
			{"Updated", created.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func editTagCategory(name string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/tagCategories/%s", cfg.APIURL, url.PathEscape(name)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	tc, err := decodeJSON[types.TagCategory](resp)
	if err != nil {
		return err
	}

	editable := editableTagCategory{
		Name:        tc.Name,
		Description: tc.Description,
		Color:       tc.Color,
	}

	var edited editableTagCategory
	changed, err := openEditor(editable, &edited)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Println("No changes.")
		return nil
	}

	updated := types.TagCategory{
		Name:        edited.Name,
		Description: edited.Description,
		Color:       edited.Color,
	}
	body, err := json.Marshal(updated)
	if err != nil {
		return fmt.Errorf("marshal tag category: %w", err)
	}
	putResp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/tagCategories/%s", cfg.APIURL, url.PathEscape(name)), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = putResp.Body.Close() }()
	if err := checkStatus(putResp); err != nil {
		return err
	}
	result, err := decodeJSON[types.TagCategory](putResp)
	if err != nil {
		return err
	}
	return printResource(result, func() [][2]string {
		return [][2]string{
			{"Name", result.Name},
			{"Description", result.Description},
			{"Color", result.Color},
			{"Created", result.CreatedAt.Format(time.RFC3339)},
			{"Updated", result.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func deleteTagCategory(name string) error {
	resp, err := doRequest(cfg, http.MethodDelete, fmt.Sprintf("%s/api/v1/tagCategories/%s", cfg.APIURL, url.PathEscape(name)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	fmt.Printf("tagcategory/%s deleted\n", name)
	return nil
}
