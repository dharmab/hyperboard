package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"uuid"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrintJSON(t *testing.T) {
	tag := types.Tag{Name: "test-tag", Description: "A test tag"}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printErr := printJSON(tag)
	require.NoError(t, printErr)
	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	var decoded types.Tag
	unmarshalErr := json.Unmarshal([]byte(output), &decoded)
	require.NoError(t, unmarshalErr)
	assert.Equal(t, "test-tag", decoded.Name)
}

func TestPrintYAML(t *testing.T) {
	tag := types.Tag{Name: "yaml-tag", Description: "A YAML tag"}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printErr := printYAML(tag)
	require.NoError(t, printErr)
	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	var decoded types.Tag
	unmarshalErr := yaml.Unmarshal([]byte(output), &decoded)
	require.NoError(t, unmarshalErr)
	assert.Equal(t, "yaml-tag", decoded.Name)
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

	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "test")
}

func TestPrintListTable(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	headers := []string{"ID", "Name"}
	rows := [][]string{
		{uuid.NewV4().String(), "tag-1"},
		{uuid.NewV4().String(), "tag-2"},
	}
	printListTable(headers, rows)

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	assert.Contains(t, output, "tag-1")
	assert.Contains(t, output, "tag-2")
}
