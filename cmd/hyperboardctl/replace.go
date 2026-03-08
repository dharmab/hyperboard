package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gabriel-vasile/mimetype"
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

	resp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/posts/%s/content", cfg.APIURL, url.PathEscape(id)), "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	post, err := decodeJSON[types.Post](resp)
	if err != nil {
		return err
	}
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

	resp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/posts/%s/thumbnail", cfg.APIURL, url.PathEscape(id)), "application/octet-stream", bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	post, err := decodeJSON[types.Post](resp)
	if err != nil {
		return err
	}
	return printResource(post, func() [][2]string {
		return postTableRows(post)
	})
}
