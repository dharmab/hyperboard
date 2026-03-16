package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"slices"
	"strings"
)

// defaultColor is the CSS color value used when no category color is available.
const defaultColor = "var(--base03)"

// quickTagEmoji is the emoji displayed next to the file size when a post has the quick-tag.
const quickTagEmoji = "⭐"

// mediaPath extracts the URL path from a raw URL string, stripping scheme and host.
func mediaPath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return strings.TrimRight(u.Path, "/")
}

// templateFuncs returns the FuncMap of custom template functions for HTML rendering.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"deref": func(s *string) string {
			if s == nil {
				return ""
			}
			return *s
		},
		"deref_int": func(i *int) int {
			if i == nil {
				return 0
			}
			return *i
		},
		"deref_strings": func(s *[]string) []string {
			if s == nil {
				return nil
			}
			return *s
		},
		"join_strings": func(s []string, sep string) string {
			return strings.Join(s, sep)
		},
		"catColor": func(colors map[string]string, cat *string) string {
			if cat == nil || colors == nil {
				return defaultColor
			}
			if c, ok := colors[*cat]; ok {
				return c
			}
			return defaultColor
		},
		"tagColor": func(colors *map[string]string, tag string) string {
			if colors == nil {
				return defaultColor
			}
			if c, ok := (*colors)[tag]; ok {
				return c
			}
			return defaultColor
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
		"toJSON": func(v any) string {
			b, err := json.Marshal(v)
			if err != nil {
				return "[]"
			}
			return string(b)
		},
		"hasTag":        slices.Contains[[]string, string],
		"quickTagEmoji": func() string { return quickTagEmoji },
		"hasPrefix":     strings.HasPrefix,
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
