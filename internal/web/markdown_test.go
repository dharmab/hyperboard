package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderMarkdown(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		contains string
	}{
		{"bold", "**bold**", "<strong>bold</strong>"},
		{"link", "[link](http://example.com)", `<a href="http://example.com"`},
		{"paragraph", "hello", "<p>hello</p>"},
		{"empty input", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderedHTML := string(renderMarkdown(tt.input))
			assert.Contains(t, renderedHTML, tt.contains, "renderMarkdown(%q)", tt.input)
		})
	}
}

func TestRenderMarkdown_XSS(t *testing.T) {
	t.Parallel()
	renderedHTML := string(renderMarkdown(`<script>alert("xss")</script>`))
	assert.NotContains(t, renderedHTML, "<script>", "renderMarkdown should strip script tags")
}
