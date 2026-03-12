package tagcategories

import (
	"context"
	"fmt"
	"time"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/spf13/cobra"
)

// editableTagCategory is a YAML-serializable subset of tag category fields for interactive editing.
type editableTagCategory struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Color       string `yaml:"color"`
}

// Register adds tag category CRUD subcommands to the CLI application.
func Register(app *cli.App) {
	getTagCategoryCmd := &cobra.Command{
		Use:   "tagcategory [name]",
		Short: "Get a tag category by name, or list all tag categories",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getTagCategory(app, args[0])
			}
			return listTagCategories(app)
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
			return createTagCategory(app, args[0], tc)
		},
	}
	createTagCategoryCmd.Flags().StringVar(&tcDescription, "description", "", "Tag category description")
	createTagCategoryCmd.Flags().StringVar(&tcColor, "color", "#888888", "Tag category color (hex)")

	editTagCategoryCmd := &cobra.Command{
		Use:   "tagcategory <name>",
		Short: "Edit a tag category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return editTagCategory(app, args[0])
		},
	}

	deleteTagCategoryCmd := &cobra.Command{
		Use:   "tagcategory <name>",
		Short: "Delete a tag category",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteTagCategory(app, args[0])
		},
	}

	app.GetCmd.AddCommand(getTagCategoryCmd)
	app.CreateCmd.AddCommand(createTagCategoryCmd)
	app.EditCmd.AddCommand(editTagCategoryCmd)
	app.DeleteCmd.AddCommand(deleteTagCategoryCmd)
}

func getTagCategory(app *cli.App, name string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.GetTagCategoryWithResponse(context.TODO(), name)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	tc := *resp.JSON200
	return app.PrintResource(tc, func() [][2]string {
		return [][2]string{
			{"Name", tc.Name},
			{"Description", tc.Description},
			{"Color", tc.Color},
			{"Created", tc.CreatedAt.Format(time.RFC3339)},
			{"Updated", tc.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func listTagCategories(app *cli.App) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	tcs, err := fetchAllTagCategories(c)
	if err != nil {
		return err
	}
	return app.PrintList(tcs, []string{"NAME", "DESCRIPTION", "COLOR"}, func() [][]string {
		rows := make([][]string, 0, len(tcs))
		for _, tc := range tcs {
			rows = append(rows, []string{tc.Name, tc.Description, tc.Color})
		}
		return rows
	})
}

func fetchAllTagCategories(c *client.ClientWithResponses) ([]types.TagCategory, error) {
	var all []types.TagCategory
	var cursor *string
	for {
		limit := 1000
		params := &client.GetTagCategoriesParams{Limit: &limit, Cursor: cursor}
		resp, err := c.GetTagCategoriesWithResponse(context.TODO(), params)
		if err != nil {
			return nil, err
		}
		if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
			return nil, err
		}
		if resp.JSON200.Items != nil {
			all = append(all, *resp.JSON200.Items...)
		}
		if resp.JSON200.Cursor == nil || *resp.JSON200.Cursor == "" {
			break
		}
		cursor = resp.JSON200.Cursor
	}
	return all, nil
}

func createTagCategory(app *cli.App, name string, tc types.TagCategory) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.PutTagCategoryWithResponse(context.TODO(), name, tc)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	created := *resp.JSON200
	return app.PrintResource(created, func() [][2]string {
		return [][2]string{
			{"Name", created.Name},
			{"Description", created.Description},
			{"Color", created.Color},
			{"Created", created.CreatedAt.Format(time.RFC3339)},
			{"Updated", created.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func editTagCategory(app *cli.App, name string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.GetTagCategoryWithResponse(context.TODO(), name)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	tc := *resp.JSON200

	editable := editableTagCategory{
		Name:        tc.Name,
		Description: tc.Description,
		Color:       tc.Color,
	}

	var edited editableTagCategory
	changed, err := cli.OpenEditor(editable, &edited)
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
	putResp, err := c.PutTagCategoryWithResponse(context.TODO(), name, updated)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(putResp.StatusCode(), putResp.Body); err != nil {
		return err
	}
	result := *putResp.JSON200
	return app.PrintResource(result, func() [][2]string {
		return [][2]string{
			{"Name", result.Name},
			{"Description", result.Description},
			{"Color", result.Color},
			{"Created", result.CreatedAt.Format(time.RFC3339)},
			{"Updated", result.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func deleteTagCategory(app *cli.App, name string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.DeleteTagCategoryWithResponse(context.TODO(), name)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	fmt.Printf("tagcategory/%s deleted\n", name)
	return nil
}
