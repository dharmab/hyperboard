package notes

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dharmab/hyperboard/internal/cli"
	"github.com/dharmab/hyperboard/pkg/client"
	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/spf13/cobra"
)

// editableNote is a YAML-serializable subset of note fields for interactive editing.
type editableNote struct {
	Title   string `yaml:"title"`
	Content string `yaml:"content"`
}

// Register adds note CRUD subcommands to the CLI application.
func Register(app *cli.App) {
	getNoteCmd := &cobra.Command{
		Use:   "note [id]",
		Short: "Get a note by ID, or list all notes",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				return getNote(cmd.Context(), app, args[0])
			}
			return listNotes(cmd.Context(), app)
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

			return createNote(cmd.Context(), app, title, content)
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
			return editNote(cmd.Context(), app, args[0])
		},
	}

	deleteNoteCmd := &cobra.Command{
		Use:   "note <id>",
		Short: "Delete a note",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteNote(cmd.Context(), app, args[0])
		},
	}

	app.GetCmd.AddCommand(getNoteCmd)
	app.CreateCmd.AddCommand(createNoteCmd)
	app.EditCmd.AddCommand(editNoteCmd)
	app.DeleteCmd.AddCommand(deleteNoteCmd)
}

func getNote(ctx context.Context, app *cli.App, id string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	noteID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.GetNoteWithResponse(ctx, types.ID(noteID))
	if err != nil {
		return err
	}
	if err := cli.CheckJSONResponse(resp.StatusCode(), resp.Body, http.StatusOK, resp.JSON200); err != nil {
		return fmt.Errorf("get note/%s: %w", id, err)
	}
	note := *resp.JSON200
	return app.PrintResource(note, func() [][2]string {
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

func listNotes(ctx context.Context, app *cli.App) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.GetNotesWithResponse(ctx)
	if err != nil {
		return err
	}
	if err := cli.CheckJSONResponse(resp.StatusCode(), resp.Body, http.StatusOK, resp.JSON200); err != nil {
		return fmt.Errorf("list notes: %w", err)
	}
	var notes []types.Note
	if resp.JSON200.Items != nil {
		notes = *resp.JSON200.Items
	}
	return app.PrintList(notes, []string{"ID", "TITLE", "CREATED"}, func() [][]string {
		rows := make([][]string, 0, len(notes))
		for _, n := range notes {
			rows = append(rows, []string{n.ID.String(), n.Title, n.CreatedAt.Format(time.RFC3339)})
		}
		return rows
	})
}

func createNote(ctx context.Context, app *cli.App, title, content string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	resp, err := c.CreateNoteWithResponse(ctx, client.CreateNoteJSONRequestBody{
		Title:   title,
		Content: content,
	})
	if err != nil {
		return err
	}
	if err := cli.CheckJSONResponse(resp.StatusCode(), resp.Body, http.StatusCreated, resp.JSON201); err != nil {
		return fmt.Errorf("create note: %w", err)
	}
	note := *resp.JSON201
	return app.PrintResource(note, func() [][2]string {
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

func editNote(ctx context.Context, app *cli.App, id string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	noteID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.GetNoteWithResponse(ctx, types.ID(noteID))
	if err != nil {
		return err
	}
	if err := cli.CheckJSONResponse(resp.StatusCode(), resp.Body, http.StatusOK, resp.JSON200); err != nil {
		return fmt.Errorf("get note/%s for editing: %w", id, err)
	}
	note := *resp.JSON200

	editable := editableNote{
		Title:   note.Title,
		Content: note.Content,
	}

	var edited editableNote
	changed, err := cli.OpenEditor(editable, &edited)
	if err != nil {
		return err
	}
	if !changed {
		fmt.Println("No changes.")
		return nil
	}

	putResp, err := c.PutNoteWithResponse(ctx, types.ID(noteID), client.PutNoteJSONRequestBody{
		Title:   edited.Title,
		Content: edited.Content,
	})
	if err != nil {
		return err
	}
	if err := cli.CheckJSONResponse(putResp.StatusCode(), putResp.Body, http.StatusOK, putResp.JSON200); err != nil {
		return fmt.Errorf("update note/%s: %w", id, err)
	}
	result := *putResp.JSON200
	return app.PrintResource(result, func() [][2]string {
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

func deleteNote(ctx context.Context, app *cli.App, id string) error {
	c, err := app.NewClient()
	if err != nil {
		return err
	}
	noteID, err := cli.ParseID(id)
	if err != nil {
		return err
	}
	resp, err := c.DeleteNoteWithResponse(ctx, types.ID(noteID))
	if err != nil {
		return err
	}
	if err := cli.CheckResponseStatus(resp.StatusCode(), resp.Body, http.StatusNoContent); err != nil {
		return fmt.Errorf("delete note/%s: %w", id, err)
	}
	fmt.Printf("note/%s deleted\n", id)
	return nil
}
