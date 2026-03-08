package main

import (
	"html/template"
	"net/url"
	"strings"
)

func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"catColor": func(colors map[string]string, cat *string) string {
			if cat == nil || colors == nil {
				return "var(--base03)"
			}
			if c, ok := colors[*cat]; ok {
				return c
			}
			return "var(--base03)"
		},
		"not": func(b bool) bool { return !b },
		"mediaUrl": func(rawURL string) string {
			u, err := url.Parse(rawURL)
			if err != nil {
				return rawURL
			}
			// Strip the scheme+host, keep the path: /bucket/key → /media/bucket/key
			return "/media" + strings.TrimRight(u.Path, "/")
		},
	}
}
