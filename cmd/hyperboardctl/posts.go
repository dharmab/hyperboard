package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gabriel-vasile/mimetype"
	"github.com/spf13/cobra"
)

type editablePost struct {
	Tags []string `yaml:"tags"`
	Note string   `yaml:"note"`
}

func init() {
	var searchQuery string

	getPostCmd := &cobra.Command{
		Use:   "post [id]",
		Short: "Get a post by ID or search posts",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getPost(args[0])
			}
			if searchQuery != "" {
				return searchPosts(searchQuery)
			}
			return fmt.Errorf("either provide a post ID or use --search")
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
			return createPost(postFile, postTags, postNote)
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
			return editPost(args[0])
		},
	}

	deletePostCmd := &cobra.Command{
		Use:   "post <id>",
		Short: "Delete a post",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deletePost(args[0])
		},
	}

	getCmd.AddCommand(getPostCmd)
	createCmd.AddCommand(createPostCmd)
	editCmd.AddCommand(editPostCmd)
	deleteCmd.AddCommand(deletePostCmd)
}

func postTableRows(post types.Post) [][2]string {
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

func getPost(id string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/posts/%s", cfg.APIURL, url.PathEscape(id)), "", nil)
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

func searchPosts(query string) error {
	params := url.Values{}
	params.Set("search", query)
	params.Set("sort", "recent")
	posts, err := fetchAll[types.Post](cfg, cfg.APIURL+"/api/v1/posts", params)
	if err != nil {
		return err
	}
	return printList(posts, []string{"ID", "MIME", "TAGS", "NOTE", "CREATED"}, func() [][]string {
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

func createPost(filePath, tagsCSV, note string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read %s: %w", filePath, err)
	}

	detected := mimetype.Detect(data)
	mimeStr := detected.String()
	if !strings.HasPrefix(mimeStr, "image/") && !strings.HasPrefix(mimeStr, "video/") {
		return fmt.Errorf("unsupported file type: %s", mimeStr)
	}

	resp, err := doRequest(cfg, http.MethodPost, cfg.APIURL+"/api/v1/upload", mimeStr, bytes.NewReader(data))
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

	if tagsCSV != "" || note != "" {
		if tagsCSV != "" {
			post.Tags = strings.Split(tagsCSV, ",")
		}
		if note != "" {
			post.Note = note
		}
		body, err := json.Marshal(post)
		if err != nil {
			return fmt.Errorf("marshal post update: %w", err)
		}
		putResp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/posts/%s", cfg.APIURL, post.ID), "application/json", bytes.NewReader(body))
		if err != nil {
			return err
		}
		defer func() { _ = putResp.Body.Close() }()
		if err := checkStatus(putResp); err != nil {
			return err
		}
		post, err = decodeJSON[types.Post](putResp)
		if err != nil {
			return err
		}
	}

	return printResource(post, func() [][2]string {
		return postTableRows(post)
	})
}

func editPost(id string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/posts/%s", cfg.APIURL, url.PathEscape(id)), "", nil)
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

	editable := editablePost{
		Tags: post.Tags,
		Note: post.Note,
	}

	var edited editablePost
	changed, err := openEditor(editable, &edited)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Println("No changes.")
		return nil
	}

	post.Tags = edited.Tags
	post.Note = edited.Note
	body, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("marshal post: %w", err)
	}
	putResp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/posts/%s", cfg.APIURL, url.PathEscape(id)), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = putResp.Body.Close() }()
	if err := checkStatus(putResp); err != nil {
		return err
	}
	result, err := decodeJSON[types.Post](putResp)
	if err != nil {
		return err
	}
	return printResource(result, func() [][2]string {
		return postTableRows(result)
	})
}

func deletePost(id string) error {
	resp, err := doRequest(cfg, http.MethodDelete, fmt.Sprintf("%s/api/v1/posts/%s", cfg.APIURL, url.PathEscape(id)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	fmt.Printf("post/%s deleted\n", id)
	return nil
}
