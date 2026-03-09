package main

import (
	"testing"
)

func TestMediaPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"full URL", "http://storage.example.com/bucket/key/file.webp", "/bucket/key/file.webp"},
		{"path only", "/bucket/key/file.webp", "/bucket/key/file.webp"},
		{"trailing slash stripped", "http://storage.example.com/bucket/key/", "/bucket/key"},
		{"unparseable returned as-is", "://bad", "://bad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := mediaPath(tt.input); got != tt.expect {
				t.Errorf("mediaPath(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	formatSize := funcs["formatSize"].(func(int64) string)

	tests := []struct {
		name   string
		input  int64
		expect string
	}{
		{"zero", 0, "0 B"},
		{"bytes", 1023, "1023 B"},
		{"kilobytes", 1024, "1.0 KB"},
		{"megabytes", 1 << 20, "1.0 MB"},
		{"gigabytes", 1 << 30, "1.0 GB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := formatSize(tt.input); got != tt.expect {
				t.Errorf("formatSize(%d) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestMediaUrl(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	mediaUrl := funcs["mediaUrl"].(func(string) string)

	tests := []struct {
		name   string
		input  string
		expect string
	}{
		{"full URL", "http://storage.example.com/bucket/key.webp", "/media/bucket/key.webp"},
		{"trailing slash stripped", "http://storage.example.com/bucket/key/", "/media/bucket/key"},
		{"unparseable returned as-is", "://bad", "://bad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := mediaUrl(tt.input); got != tt.expect {
				t.Errorf("mediaUrl(%q) = %q, want %q", tt.input, got, tt.expect)
			}
		})
	}
}

func TestCatColor(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	catColor := funcs["catColor"].(func(map[string]string, *string) string)
	defaultColor := "var(--base03)"

	t.Run("nil cat", func(t *testing.T) {
		t.Parallel()
		if got := catColor(map[string]string{"a": "#fff"}, nil); got != defaultColor {
			t.Errorf("catColor(colors, nil) = %q, want %q", got, defaultColor)
		}
	})
	t.Run("nil colors", func(t *testing.T) {
		t.Parallel()
		cat := "a"
		if got := catColor(nil, &cat); got != defaultColor {
			t.Errorf("catColor(nil, &cat) = %q, want %q", got, defaultColor)
		}
	})
	t.Run("missing key", func(t *testing.T) {
		t.Parallel()
		cat := "missing"
		if got := catColor(map[string]string{"a": "#fff"}, &cat); got != defaultColor {
			t.Errorf("catColor(colors, &missing) = %q, want %q", got, defaultColor)
		}
	})
	t.Run("found key", func(t *testing.T) {
		t.Parallel()
		cat := "a"
		if got := catColor(map[string]string{"a": "#fff"}, &cat); got != "#fff" {
			t.Errorf("catColor(colors, &a) = %q, want %q", got, "#fff")
		}
	})
}

func TestDeref(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	deref := funcs["deref"].(func(*string) string)

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		if got := deref(nil); got != "" {
			t.Errorf("deref(nil) = %q, want empty", got)
		}
	})
	t.Run("non-nil", func(t *testing.T) {
		t.Parallel()
		s := "hello"
		if got := deref(&s); got != "hello" {
			t.Errorf("deref(&s) = %q, want %q", got, "hello")
		}
	})
}

func TestDerefInt(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	derefInt := funcs["deref_int"].(func(*int) int)

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		if got := derefInt(nil); got != 0 {
			t.Errorf("deref_int(nil) = %d, want 0", got)
		}
	})
	t.Run("non-nil", func(t *testing.T) {
		t.Parallel()
		i := 42
		if got := derefInt(&i); got != 42 {
			t.Errorf("deref_int(&i) = %d, want 42", got)
		}
	})
}

func TestDerefStrings(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	derefStrings := funcs["deref_strings"].(func(*[]string) []string)

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		if got := derefStrings(nil); got != nil {
			t.Errorf("deref_strings(nil) = %v, want nil", got)
		}
	})
	t.Run("non-nil", func(t *testing.T) {
		t.Parallel()
		s := []string{"a", "b"}
		got := derefStrings(&s)
		if len(got) != 2 || got[0] != "a" || got[1] != "b" {
			t.Errorf("deref_strings(&s) = %v, want [a b]", got)
		}
	})
}

func TestNot(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	not := funcs["not"].(func(bool) bool)

	if not(true) {
		t.Error("not(true) = true, want false")
	}
	if !not(false) {
		t.Error("not(false) = false, want true")
	}
}

func TestJoinStrings(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	joinStrings := funcs["join_strings"].(func([]string, string) string)

	if got := joinStrings([]string{"a", "b", "c"}, ", "); got != "a, b, c" {
		t.Errorf("join_strings = %q, want %q", got, "a, b, c")
	}
}
