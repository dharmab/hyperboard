package web

import (
	"bytes"
	"html/template"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
)

// renderMarkdown converts a Markdown string to sanitized HTML.
func renderMarkdown(src string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(src), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	sanitized := bluemonday.UGCPolicy().SanitizeBytes(buf.Bytes())
	return template.HTML(sanitized)
}
