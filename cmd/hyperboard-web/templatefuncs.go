package main

import (
	"fmt"
	"html/template"
	"net/url"
	"strings"
)

func mediaPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return strings.TrimRight(u.Path, "/")
}

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
		"formatSize": func(bytes int64) string {
			switch {
			case bytes >= 1<<30:
				return fmt.Sprintf("%.1f GB", float64(bytes)/float64(1<<30))
			case bytes >= 1<<20:
				return fmt.Sprintf("%.1f MB", float64(bytes)/float64(1<<20))
			case bytes >= 1<<10:
				return fmt.Sprintf("%.1f KB", float64(bytes)/float64(1<<10))
			default:
				return fmt.Sprintf("%d B", bytes)
			}
		},
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
