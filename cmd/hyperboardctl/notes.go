package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/spf13/cobra"
)

type editableNote struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

func init() {
	getNoteCmd := &cobra.Command{
		Use:   "note [id]",
		Short: "Get a note by ID, or list all notes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getNote(args[0])
			}
			return listNotes()
		},
	}

	var noteFile string
	var noteTitle string
	var noteContent string

	createNoteCmd := &cobra.Command{
		Use:   "note",
		Short: "Create a note",
		RunE: func(cmd *cobra.Command, args []string) error {
			var title, content string

			if noteFile != "" {
				data, err := os.ReadFile(noteFile)
				if err != nil {
					return fmt.Errorf("read file: %w", err)
				}
				content = string(data)
				if noteTitle != "" {
					title = noteTitle
				} else {
					title = strings.TrimSuffix(filepath.Base(noteFile), filepath.Ext(noteFile))
				}
			} else if noteTitle != "" {
				title = noteTitle
				content = noteContent
			} else {
				return fmt.Errorf("either -f/--file or --title is required")
			}

			return createNote(title, content)
		},
	}
	createNoteCmd.Flags().StringVarP(&noteFile, "file", "f", "", "Path to markdown file")
	createNoteCmd.Flags().StringVar(&noteTitle, "title", "", "Note title")
	createNoteCmd.Flags().StringVar(&noteContent, "content", "", "Note content")

	editNoteCmd := &cobra.Command{
		Use:   "note <id>",
		Short: "Edit a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return editNote(args[0])
		},
	}

	deleteNoteCmd := &cobra.Command{
		Use:   "note <id>",
		Short: "Delete a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteNote(args[0])
		},
	}

	getCmd.AddCommand(getNoteCmd)
	createCmd.AddCommand(createNoteCmd)
	editCmd.AddCommand(editNoteCmd)
	deleteCmd.AddCommand(deleteNoteCmd)
}

func getNote(id string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/notes/%s", cfg.APIURL, url.PathEscape(id)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	note, err := decodeJSON[types.Note](resp)
	if err != nil {
		return err
	}
	return printResource(note, func() [][2]string {
		content := note.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		return [][2]string{
			{"ID", note.ID.String()},
			{"Title", note.Title},
			{"Content", content},
			{"Created", note.CreatedAt.Format(time.RFC3339)},
			{"Updated", note.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func listNotes() error {
	resp, err := doRequest(cfg, http.MethodGet, cfg.APIURL+"/api/v1/notes", "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	page, err := decodeJSON[listResponse[types.Note]](resp)
	if err != nil {
		return err
	}
	var notes []types.Note
	if page.Items != nil {
		notes = *page.Items
	}
	return printList(notes, []string{"ID", "TITLE", "CREATED"}, func() [][]string {
		rows := make([][]string, 0, len(notes))
		for _, n := range notes {
			rows = append(rows, []string{n.ID.String(), n.Title, n.CreatedAt.Format(time.RFC3339)})
		}
		return rows
	})
}

func createNote(title, content string) error {
	reqBody := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{Title: title, Content: content}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal note: %w", err)
	}
	resp, err := doRequest(cfg, http.MethodPost, cfg.APIURL+"/api/v1/notes", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	note, err := decodeJSON[types.Note](resp)
	if err != nil {
		return err
	}
	return printResource(note, func() [][2]string {
		content := note.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		return [][2]string{
			{"ID", note.ID.String()},
			{"Title", note.Title},
			{"Content", content},
			{"Created", note.CreatedAt.Format(time.RFC3339)},
			{"Updated", note.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func editNote(id string) error {
	resp, err := doRequest(cfg, http.MethodGet, fmt.Sprintf("%s/api/v1/notes/%s", cfg.APIURL, url.PathEscape(id)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	note, err := decodeJSON[types.Note](resp)
	if err != nil {
		return err
	}

	editable := editableNote{
		Title:   note.Title,
		Content: note.Content,
	}

	var edited editableNote
	changed, err := openEditor(editable, &edited)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Println("No changes.")
		return nil
	}

	reqBody := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{Title: edited.Title, Content: edited.Content}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal note: %w", err)
	}
	putResp, err := doRequest(cfg, http.MethodPut, fmt.Sprintf("%s/api/v1/notes/%s", cfg.APIURL, url.PathEscape(id)), "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer func() { _ = putResp.Body.Close() }()
	if err := checkStatus(putResp); err != nil {
		return err
	}
	result, err := decodeJSON[types.Note](putResp)
	if err != nil {
		return err
	}
	return printResource(result, func() [][2]string {
		content := result.Content
		if len(content) > 80 {
			content = content[:80] + "..."
		}
		return [][2]string{
			{"ID", result.ID.String()},
			{"Title", result.Title},
			{"Content", content},
			{"Created", result.CreatedAt.Format(time.RFC3339)},
			{"Updated", result.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func deleteNote(id string) error {
	resp, err := doRequest(cfg, http.MethodDelete, fmt.Sprintf("%s/api/v1/notes/%s", cfg.APIURL, url.PathEscape(id)), "", nil)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if err := checkStatus(resp); err != nil {
		return err
	}
	fmt.Printf("note/%s deleted\n", id)
	return nil
}
