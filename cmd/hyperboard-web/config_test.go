package main

import (
	"testing"
)

func TestParseTagFilters(t *testing.T) {
	t.Parallel()

	t.Run("valid JSON", func(t *testing.T) {
		t.Parallel()
		input := `[{"label":"Rating","tags":["rating:safe","rating:questionable"]},{"label":"No AI","tags":["-ai_generated"]}]`
		filters := parseTagFilters(input)
		if len(filters) != 2 {
			t.Fatalf("got %d filters, want 2", len(filters))
		}
		if filters[0].Label != "Rating" {
			t.Errorf("filters[0].Label = %q, want %q", filters[0].Label, "Rating")
		}
		if len(filters[0].Tags) != 2 {
			t.Errorf("filters[0].Tags length = %d, want 2", len(filters[0].Tags))
		}
		if filters[1].Label != "No AI" {
			t.Errorf("filters[1].Label = %q, want %q", filters[1].Label, "No AI")
		}
		if len(filters[1].Tags) != 1 || filters[1].Tags[0] != "-ai_generated" {
			t.Errorf("filters[1].Tags = %v, want [\"-ai_generated\"]", filters[1].Tags)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		t.Parallel()
		filters := parseTagFilters("")
		if filters != nil {
			t.Errorf("got %v, want nil", filters)
		}
	})

	t.Run("malformed JSON", func(t *testing.T) {
		t.Parallel()
		filters := parseTagFilters("{bad json")
		if filters != nil {
			t.Errorf("got %v, want nil", filters)
		}
	})
}
