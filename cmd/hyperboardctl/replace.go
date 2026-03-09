package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

func init() {
	var contentFile string

	replaceContentCmd := &cobra.Command{
		Use:   "content <id>",
		Short: "Replace a post's content from file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return replaceContent(args[0], contentFile)
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
			return replaceThumbnail(args[0], thumbnailFile)
		},
	}
	replaceThumbnailCmd.Flags().StringVarP(&thumbnailFile, "file", "f", "", "Path to image file (required)")
	_ = replaceThumbnailCmd.MarkFlagRequired("file")

	replaceCmd.AddCommand(replaceContentCmd, replaceThumbnailCmd)
}

func replaceContent(id, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	detected := mimetype.Detect(data)
	mimeStr := detected.String()
	if !strings.HasPrefix(mimeStr, "image/") && !strings.HasPrefix(mimeStr, "video/") {
		return fmt.Errorf("unsupported file type: %s", mimeStr)
	}

	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	postID, err := parseID(id)
	if err != nil {
		return err
	}
	resp, err := c.ReplacePostContentWithBodyWithResponse(context.TODO(), postID, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	post := *resp.JSON200
	return printResource(post, func() [][2]string {
		return postTableRows(post)
	})
}

func replaceThumbnail(id, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	detected := mimetype.Detect(data)
	mimeStr := detected.String()
	if !strings.HasPrefix(mimeStr, "image/") {
		return fmt.Errorf("unsupported file type (must be image/*): %s", mimeStr)
	}

	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	postID, err := parseID(id)
	if err != nil {
		return err
	}
	resp, err := c.ReplacePostThumbnailWithBodyWithResponse(context.TODO(), postID, "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	post := *resp.JSON200
	return printResource(post, func() [][2]string {
		return postTableRows(post)
	})
}

func parseID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return id, fmt.Errorf("invalid ID %q: %w", s, err)
	}
	return id, nil
}
