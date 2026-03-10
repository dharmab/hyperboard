package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/pkg/client"
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
				return errors.New("either -f/--file or --title is required")
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
	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	noteID, err := parseID(id)
	if err != nil {
		return err
	}
	resp, err := c.GetNoteWithResponse(context.TODO(), noteID)
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	note := *resp.JSON200
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
	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	resp, err := c.GetNotesWithResponse(context.TODO())
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	var notes []types.Note
	if resp.JSON200 != nil && resp.JSON200.Items != nil {
		notes = *resp.JSON200.Items
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
	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	resp, err := c.CreateNoteWithResponse(context.TODO(), client.CreateNoteJSONRequestBody{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	note := *resp.JSON201
	return printResource(note, func() [][2]string {
		c := note.Content
		if len(c) > 80 {
			c = c[:80] + "..."
		}
		return [][2]string{
			{"ID", note.ID.String()},
			{"Title", note.Title},
			{"Content", c},
			{"Created", note.CreatedAt.Format(time.RFC3339)},
			{"Updated", note.UpdatedAt.Format(time.RFC3339)},
		}
	})
}

func editNote(id string) error {
	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	noteID, err := parseID(id)
	if err != nil {
		return err
	}
	resp, err := c.GetNoteWithResponse(context.TODO(), noteID)
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	note := *resp.JSON200

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

	putResp, err := c.PutNoteWithResponse(context.TODO(), noteID, client.PutNoteJSONRequestBody{
		Title:   edited.Title,
		Content: edited.Content,
	})
	if err != nil {
		return err
	}
	if err := checkResponse(putResp.StatusCode(), putResp.Body); err != nil {
		return err
	}
	result := *putResp.JSON200
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
	c, err := newClient(cfg)
	if err != nil {
		return err
	}
	noteID, err := parseID(id)
	if err != nil {
		return err
	}
	resp, err := c.DeleteNoteWithResponse(context.TODO(), noteID)
	if err != nil {
		return err
	}
	if err := checkResponse(resp.StatusCode(), resp.Body); err != nil {
		return err
	}
	fmt.Printf("note/%s deleted\n", id)
	return nil
}
