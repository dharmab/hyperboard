package api

import "testing"

func TestParseLimit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  *int
		expect int
	}{
		{"nil defaults to MaxLimit", nil, MaxLimit},
		{"zero defaults to MaxLimit", new(0), MaxLimit},
		{"negative defaults to MaxLimit", new(-1), MaxLimit},
		{"within range", new(10), 10},
		{"at max", new(MaxLimit), MaxLimit},
		{"exceeds max capped", new(MaxLimit + 100), MaxLimit},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := parseLimit(tt.input); got != tt.expect {
				t.Errorf("parseLimit() = %d, want %d", got, tt.expect)
			}
		})
	}
}

func TestObfuscateCursorRoundTrip(t *testing.T) {
	t.Parallel()
	original := "test-cursor-value"
	encoded := obfuscateCursor(original)
	decoded, err := deobfuscateCursor(encoded)
	if err != nil {
		t.Fatalf("deobfuscateCursor() error = %v", err)
	}
	if decoded != original {
		t.Errorf("round-trip failed: got %q, want %q", decoded, original)
	}
}

func TestDeobfuscateCursorInvalid(t *testing.T) {
	t.Parallel()
	_, err := deobfuscateCursor("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}
}

func TestPaginate(t *testing.T) {
	t.Parallel()
	t.Run("more results available", func(t *testing.T) {
		t.Parallel()
		more, cursor := paginate(11, 10, func() string { return "page-value" })
		if !more {
			t.Error("expected more = true")
		}
		if cursor == nil {
			t.Fatal("expected non-nil cursor")
		}
		decoded, err := deobfuscateCursor(*cursor)
		if err != nil {
			t.Fatalf("deobfuscateCursor() error = %v", err)
		}
		if decoded != "page-value" {
			t.Errorf("cursor decoded to %q, want %q", decoded, "page-value")
		}
	})

	t.Run("no more results", func(t *testing.T) {
		t.Parallel()
		more, cursor := paginate(10, 10, func() string { return "unused" })
		if more {
			t.Error("expected more = false")
		}
		if cursor != nil {
			t.Error("expected nil cursor")
		}
	})

	t.Run("fewer results than limit", func(t *testing.T) {
		t.Parallel()
		more, cursor := paginate(5, 10, func() string { return "unused" })
		if more {
			t.Error("expected more = false")
		}
		if cursor != nil {
			t.Error("expected nil cursor")
		}
	})
}
