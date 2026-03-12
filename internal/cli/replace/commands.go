package replace

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/internal/cli/posts"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
)

// Register adds content and thumbnail replacement subcommands to the CLI application.
func Register(app *cli.App) {
	var contentFile string

	replaceContentCmd := &cobra.Command{
		Use:   "content <id>",
		Short: "Replace a post's content from file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return replaceContent(app, args[0], contentFile)
		},
	}
	replaceContentCmd.Flags().StringVarP(&contentFile, "file", "f", "", "Path to media file (required)")
	_ = replaceContentCmd.MarkFlagRequired("file")

	var thumbnailFile string

	replaceThumbnailCmd := &cobra.Command{
		Use:   "thumbnail <id>",
		Short: "Replace a post's thumbnail from file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return replaceThumbnail(app, args[0], thumbnailFile)
		},
	}
	replaceThumbnailCmd.Flags().StringVarP(&thumbnailFile, "file", "f", "", "Path to image file (required)")
	_ = replaceThumbnailCmd.MarkFlagRequired("file")

	app.ReplaceCmd.AddCommand(replaceContentCmd, replaceThumbnailCmd)
}

func replaceContent(app *cli.App, id, filePath string) error {
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
	postID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.ReplacePostContentWithBodyWithResponse(context.TODO(), postID, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	post := *resp.JSON200
	return app.PrintResource(post, func() [][2]string {
		return posts.TableRows(post)
	})
}

func replaceThumbnail(app *cli.App, id, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	detected := mimetype.Detect(data)
	mimeStr := detected.String()
	if !strings.HasPrefix(mimeStr, "image/") {
		return fmt.Errorf("unsupported file type (must be image/*): %s", mimeStr)
	}

	c, err := app.NewClient()
	if err != nil {
		return err
	}
	postID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.ReplacePostThumbnailWithBodyWithResponse(context.TODO(), postID, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := cli.CheckResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	post := *resp.JSON200
	return app.PrintResource(post, func() [][2]string {
		return posts.TableRows(post)
	})
}
