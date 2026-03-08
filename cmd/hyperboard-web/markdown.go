package main

import (
	"bytes"
	"html/template"

	"github.com/yuin/goldmark"
)

func renderMarkdown(src string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(src), &buf); err != nil {
		return template.HTML(template.HTMLEscapeString(src))
	}
	return template.HTML(buf.String())
}
