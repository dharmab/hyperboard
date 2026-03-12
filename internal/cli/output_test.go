package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/gofrs/uuid/v5"
	"gopkg.in/yaml.v3"
)

func TestPrintJSON(t *testing.T) {
	tag := types.Tag{Name: "test-tag", Description: "A test tag"}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := printJSON(tag); err != nil {
		t.Fatalf("printJSON error: %v", err)
	}
	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	var decoded types.Tag
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("failed to parse JSON output: %v", err)
	}
	if decoded.Name != "test-tag" {
		t.Errorf("Name = %q, want %q", decoded.Name, "test-tag")
	}
}

func TestPrintYAML(t *testing.T) {
	tag := types.Tag{Name: "yaml-tag", Description: "A YAML tag"}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	if err := printYAML(tag); err != nil {
		t.Fatalf("printYAML error: %v", err)
	}
	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	var decoded types.Tag
	if err := yaml.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("failed to parse YAML output: %v", err)
	}
	if decoded.Name != "yaml-tag" {
		t.Errorf("Name = %q, want %q", decoded.Name, "yaml-tag")
	}
}

func TestPrintTable(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printTable([][2]string{{"Name", "test"}, {"Value", "123"}})

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "Name") || !strings.Contains(output, "test") {
		t.Errorf("expected table output to contain 'Name' and 'test', got: %s", output)
	}
}

func TestPrintListTable(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	headers := []string{"ID", "Name"}
	rows := [][]string{
		{uuid.Must(uuid.NewV4()).String(), "tag-1"},
		{uuid.Must(uuid.NewV4()).String(), "tag-2"},
	}
	printListTable(headers, rows)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	if !strings.Contains(output, "tag-1") || !strings.Contains(output, "tag-2") {
		t.Errorf("expected list table to contain tag names, got: %s", output)
	}
}
