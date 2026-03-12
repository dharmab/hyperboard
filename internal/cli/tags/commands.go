package tags

import (
	"context"
	"fmt"
	"time"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/spf13/cobra"
)

// noCategoryLabel is the display label used when a tag has no category.
const noCategoryLabel = "(none)"

// editableTag is a YAML-serializable subset of tag fields for interactive editing.
type editableTag struct {
	Name        string  `yaml:"name"`
	Category    *string `yaml:"category,omitempty"`
	Description string  `yaml:"description"`
}

// Register adds tag CRUD subcommands to the CLI application.
func Register(app *cli.App) {
	getTagCmd := &cobra.Command{
		Use:   "tag [name]",
		Short: "Get a tag by name, or list all tags",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getTag(app, args[0])
			}
			return listTags(app)
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
			return createTag(app, args[0], tag)
		},
	}
	createTagCmd.Flags().StringVar(&tagDescription, "description", "", "Tag description")
	createTagCmd.Flags().StringVar(&tagCategory, "category", "", "Tag category name")

	editTagCmd := &cobra.Command{
		Use:   "tag <name>",
		Short: "Edit a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return editTag(app, args[0])
		},
	}

	deleteTagCmd := &cobra.Command{
		Use:   "tag <name>",
		Short: "Delete a tag",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteTag(app, args[0])
		},
	}

	app.GetCmd.AddCommand(getTagCmd)
	app.CreateCmd.AddCommand(createTagCmd)
	app.EditCmd.AddCommand(editTagCmd)
	app.DeleteCmd.AddCommand(deleteTagCmd)
}

func getTag(app *cli.App, name string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.GetTagWithResponse(context.TODO(), name)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	tag := *resp.JSON200
	return app.PrintResource(tag, func() [][2]string {
		cat := noCategoryLabel
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

func listTags(app *cli.App) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	tags, err := FetchAllTags(c)
	if err != nil {
		return err
	}
	return app.PrintList(tags, []string{"NAME", "CATEGORY", "DESCRIPTION"}, func() [][]string {
		rows := make([][]string, 0, len(tags))
		for _, t := range tags {
			cat := noCategoryLabel
			if t.Category != nil {
				cat = *t.Category
			}
			rows = append(rows, []string{t.Name, cat, t.Description})
		}
		return rows
	})
}

func FetchAllTags(c *client.ClientWithResponses) ([]types.Tag, error) {
	var all []types.Tag
	var cursor *string
	for {
		limit := 1000
		params := &client.GetTagsParams{Limit: &limit, Cursor: cursor}
		resp, err := c.GetTagsWithResponse(context.TODO(), params)
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

func createTag(app *cli.App, name string, tag types.Tag) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.PutTagWithResponse(context.TODO(), name, tag)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	created := *resp.JSON201
	return app.PrintResource(created, func() [][2]string {
		cat := noCategoryLabel
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

func editTag(app *cli.App, name string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.GetTagWithResponse(context.TODO(), name)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	tag := *resp.JSON200

	editable := editableTag{
		Name:        tag.Name,
		Category:    tag.Category,
		Description: tag.Description,
	}

	var edited editableTag
	changed, err := cli.OpenEditor(editable, &edited)
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
	putResp, err := c.PutTagWithResponse(context.TODO(), name, updated)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(putResp.StatusCode(), putResp.Body); err != nil {
		return err
	}
	result := *putResp.JSON201
	return app.PrintResource(result, func() [][2]string {
		cat := noCategoryLabel
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

func deleteTag(app *cli.App, name string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.DeleteTagWithResponse(context.TODO(), name)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	fmt.Printf("tag/%s deleted\n", name)
	return nil
}
