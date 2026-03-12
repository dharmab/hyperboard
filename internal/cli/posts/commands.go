package posts

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
)

type editablePost struct {
	Tags []string `yaml:"tags"`
	Note string   `yaml:"note"`
}

func Register(app *cli.App) {
	var searchQuery string

	getPostCmd := &cobra.Command{
		Use:   "post [id]",
		Short: "Get a post by ID or search posts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getPost(app, args[0])
			}
			if searchQuery != "" {
				return searchPosts(app, searchQuery)
			}
			return errors.New("either provide a post ID or use --search")
		},
	}
	getPostCmd.Flags().StringVar(&searchQuery, "search", "", "Search query")

	var postFile string
	var postTags string
	var postNote string

	createPostCmd := &cobra.Command{
		Use:   "post",
		Short: "Create a post by uploading a file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return createPost(app, postFile, postTags, postNote)
		},
	}
	createPostCmd.Flags().StringVarP(&postFile, "file", "f", "", "Path to media file (required)")
	_ = createPostCmd.MarkFlagRequired("file")
	createPostCmd.Flags().StringVar(&postTags, "tags", "", "Comma-separated tag names")
	createPostCmd.Flags().StringVar(&postNote, "note", "", "Note text")

	editPostCmd := &cobra.Command{
		Use:   "post <id>",
		Short: "Edit a post's metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return editPost(app, args[0])
		},
	}

	deletePostCmd := &cobra.Command{
		Use:   "post <id>",
		Short: "Delete a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deletePost(app, args[0])
		},
	}

	app.GetCmd.AddCommand(getPostCmd)
	app.CreateCmd.AddCommand(createPostCmd)
	app.EditCmd.AddCommand(editPostCmd)
	app.DeleteCmd.AddCommand(deletePostCmd)
}

func TableRows(post types.Post) [][2]string {
	return [][2]string{
		{"ID", post.ID.String()},
		{"MIME Type", post.MimeType},
		{"Content URL", post.ContentUrl},
		{"Thumbnail URL", post.ThumbnailUrl},
		{"Tags", strings.Join(post.Tags, ", ")},
		{"Note", post.Note},
		{"Created", post.CreatedAt.Format(time.RFC3339)},
		{"Updated", post.UpdatedAt.Format(time.RFC3339)},
	}
}

func getPost(app *cli.App, id string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	postID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.GetPostWithResponse(context.TODO(), postID)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	post := *resp.JSON200
	return app.PrintResource(post, func() [][2]string {
		return TableRows(post)
	})
}

func searchPosts(app *cli.App, query string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	search := query
	posts, err := fetchAllPosts(c, &client.GetPostsParams{Search: &search})
	if err != nil {
		return err
	}
	return app.PrintList(posts, []string{"ID", "MIME", "TAGS", "NOTE", "CREATED"}, func() [][]string {
		rows := make([][]string, 0, len(posts))
		for _, p := range posts {
			rows = append(rows, []string{
				p.ID.String(),
				p.MimeType,
				strings.Join(p.Tags, ","),
				p.Note,
				p.CreatedAt.Format(time.RFC3339),
			})
		}
		return rows
	})
}

func fetchAllPosts(c *client.ClientWithResponses, baseParams *client.GetPostsParams) ([]types.Post, error) {
	var all []types.Post
	params := *baseParams
	for {
		resp, err := c.GetPostsWithResponse(context.TODO(), &params)
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
		params.Cursor = resp.JSON200.Cursor
	}
	return all, nil
}

func createPost(app *cli.App, filePath, tagsCSV, note string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	detected := mimetype.Detect(data)
	mimeStr := detected.String()
	if !strings.HasPrefix(mimeStr, "image/") && !strings.HasPrefix(mimeStr, "video/") {
		return fmt.Errorf("unsupported file type: %s", mimeStr)
	}

	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.UploadPostWithBodyWithResponse(context.TODO(), &client.UploadPostParams{}, mimeStr, bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	post := *resp.JSON201

	if tagsCSV != "" || note != "" {
		if tagsCSV != "" {
			post.Tags = strings.Split(tagsCSV, ",")
		}
		if note != "" {
			post.Note = note
		}
		putResp, err := c.PutPostWithResponse(context.TODO(), post.ID, post)
		if err != nil {
			return err
		}
		if err := cli.CheckResponse(putResp.StatusCode(), putResp.Body); err != nil {
			return err
		}
		post = *putResp.JSON200
	}

	return app.PrintResource(post, func() [][2]string {
		return TableRows(post)
	})
}

func editPost(app *cli.App, id string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	postID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.GetPostWithResponse(context.TODO(), postID)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	post := *resp.JSON200

	editable := editablePost{
		Tags: post.Tags,
		Note: post.Note,
	}

	var edited editablePost
	changed, err := cli.OpenEditor(editable, &edited)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Println("No changes.")
		return nil
	}

	post.Tags = edited.Tags
	post.Note = edited.Note
	putResp, err := c.PutPostWithResponse(context.TODO(), postID, post)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(putResp.StatusCode(), putResp.Body); err != nil {
		return err
	}
	result := *putResp.JSON200
	return app.PrintResource(result, func() [][2]string {
		return TableRows(result)
	})
}

func deletePost(app *cli.App, id string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	postID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.DeletePostWithResponse(context.TODO(), postID)
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	fmt.Printf("post/%s deleted\n", id)
	return nil
}
