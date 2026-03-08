package main

import "html/template"

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
	}
}
