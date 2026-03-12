package web

import (
	"strings"
	"testing"
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
			got := string(renderMarkdown(tt.input))
			if !strings.Contains(got, tt.contains) {
				t.Errorf("renderMarkdown(%q) = %q, want it to contain %q", tt.input, got, tt.contains)
			}
		})
	}
}

func TestRenderMarkdown_XSS(t *testing.T) {
	t.Parallel()
	got := string(renderMarkdown(`<script>alert("xss")</script>`))
	if strings.Contains(got, "<script>") {
		t.Errorf("renderMarkdown should strip <script> tags, got %q", got)
	}
}
