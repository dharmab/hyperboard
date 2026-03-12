package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// printTable writes key-value pairs as a tab-aligned table to stdout.
func printTable(rows [][2]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, row := range rows {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", row[0], row[1])
	}
	_ = w.Flush()
}

// printListTable writes a table with headers and rows to stdout.
func printListTable(headers []string, rows [][]string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for i, h := range headers {
		if i > 0 {
			_, _ = fmt.Fprint(w, "\t")
		}
		_, _ = fmt.Fprint(w, h)
	}
	_, _ = fmt.Fprintln(w)
	for _, row := range rows {
		for i, col := range row {
			if i > 0 {
				_, _ = fmt.Fprint(w, "\t")
			}
			_, _ = fmt.Fprint(w, col)
		}
		_, _ = fmt.Fprintln(w)
	}
	_ = w.Flush()
}

// printYAML marshals a value as YAML and writes it to stdout.
func printYAML(v any) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}
	_, err = os.Stdout.Write(data)
	return err
}

// printJSON marshals a value as indented JSON and writes it to stdout.
func printJSON(v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}
	_, err = fmt.Fprintln(os.Stdout, string(data))
	return err
}

// PrintResource outputs a single resource in the configured format.
func (a *App) PrintResource(v any, tableRows func() [][2]string) error {
	switch a.OutputFormat {
	case "yaml":
		return printYAML(v)
	case "json":
		return printJSON(v)
	default:
		printTable(tableRows())
		return nil
	}
}

// PrintList outputs a list of resources in the configured format.
func (a *App) PrintList(v any, headers []string, rowFn func() [][]string) error {
	switch a.OutputFormat {
	case "yaml":
		return printYAML(v)
	case "json":
		return printJSON(v)
	default:
		printListTable(headers, rowFn())
		return nil
	}
}

func OpenEditor(v any, out any) (bool, error) {
	data, err := yaml.Marshal(v)
	if err != nil {
		return false, fmt.Errorf("marshal for editor: %w", err)
	}

	tmpFile, err := os.CreateTemp("", "hyperboardctl-*.yaml")
	if err != nil {
		return false, fmt.Errorf("create temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write(data); err != nil {
		_ = tmpFile.Close()
		return false, fmt.Errorf("write temp file: %w", err)
	}
	_ = tmpFile.Close()

	editor := os.Getenv("VISUAL")
	if editor == "" {
		editor = os.Getenv("EDITOR")
	}
	if editor == "" {
		editor = "vi"
	}

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return false, fmt.Errorf("editor: %w", err)
	}

	edited, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return false, fmt.Errorf("read edited file: %w", err)
	}

	if err := yaml.Unmarshal(edited, out); err != nil {
		return false, fmt.Errorf("parse edited file: %w", err)
	}

	return !reflect.DeepEqual(v, out), nil
}
