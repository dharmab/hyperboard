package web

import (
	"bytes"
	"testing"

	"github.com/dharmab/hyperboard/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEscapePathSegment(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "tag%2Fname%20with%20spaces%3F%23", escapePathSegment("tag/name with spaces?#"))
}

func TestResourceNamesArePathEscapedInTemplates(t *testing.T) {
	t.Parallel()

	tmpls, err := parseTemplates()
	require.NoError(t, err)

	render := func(t *testing.T, name string, data any) string {
		t.Helper()
		var output bytes.Buffer
		require.NoError(t, tmpls[name].ExecuteTemplate(&output, name, data))
		return output.String()
	}

	const (
		tagName         = "tag/name ?#"
		escapedTagName  = "tag%2Fname%20%3F%23"
		categoryName    = "category/name ?#"
		escapedCategory = "category%2Fname%20%3F%23"
	)

	t.Run("tag listing links", func(t *testing.T) {
		t.Parallel()
		category := categoryName
		postCount := 1
		body := render(t, "tags", tagsData{
			Tags: []types.Tag{{Name: tagName, Category: &category, PostCount: &postCount}},
		})
		assert.Contains(t, body, `href="/tags/`+escapedTagName+`"`)
		assert.Contains(t, body, `href="/tag-categories/`+escapedCategory+`"`)
		assert.Contains(t, body, `href="/?search=tag%2fname%20%3f%23"`)
	})

	t.Run("tag edit actions", func(t *testing.T) {
		t.Parallel()
		body := render(t, "tag_edit", tagEditData{CurrentName: tagName})
		assert.Contains(t, body, `action="/tags/`+escapedTagName+`"`)
		assert.Contains(t, body, `hx-delete="/tags/`+escapedTagName+`"`)
		assert.Contains(t, body, `action="/tags/`+escapedTagName+`/convert-to-alias"`)
		assert.Contains(t, body, `hx-get="/tag-suggestions?exclude=tag/name ?#"`)
	})

	t.Run("category listing link", func(t *testing.T) {
		t.Parallel()
		body := render(t, "tag_categories", tagCategoriesData{
			Categories: []types.TagCategory{{Name: categoryName}},
		})
		assert.Contains(t, body, `href="/tag-categories/`+escapedCategory+`"`)
	})

	t.Run("category edit actions", func(t *testing.T) {
		t.Parallel()
		body := render(t, "tag_category_edit", tagCategoryEditData{CurrentName: categoryName})
		assert.Contains(t, body, `action="/tag-categories/`+escapedCategory+`"`)
		assert.Contains(t, body, `hx-delete="/tag-categories/`+escapedCategory+`"`)
	})

	t.Run("post tag removal action", func(t *testing.T) {
		t.Parallel()
		body := render(t, "post", postData{Post: types.Post{Tags: []string{tagName}}})
		assert.Contains(t, body, `/tags/`+escapedTagName+`"`)
		assert.Contains(t, body, `href="/?search=tag%2fname%20%3f%23"`)
	})
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
			formatted := formatSize(tt.input)
			assert.Equal(t, tt.expect, formatted)
		})
	}
}

func TestCatColor(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	catColor := funcs["catColor"].(func(map[string]string, *string) string)
	const defaultColor = "var(--base03)"

	t.Run("nil cat", func(t *testing.T) {
		t.Parallel()
		color := catColor(map[string]string{"a": "#fff"}, nil)
		assert.Equal(t, defaultColor, color)
	})
	t.Run("nil colors", func(t *testing.T) {
		t.Parallel()
		cat := "a"
		color := catColor(nil, &cat)
		assert.Equal(t, defaultColor, color)
	})
	t.Run("missing key", func(t *testing.T) {
		t.Parallel()
		cat := "missing"
		color := catColor(map[string]string{"a": "#fff"}, &cat)
		assert.Equal(t, defaultColor, color)
	})
	t.Run("found key", func(t *testing.T) {
		t.Parallel()
		cat := "a"
		color := catColor(map[string]string{"a": "#fff"}, &cat)
		assert.Equal(t, "#fff", color)
	})
}

func TestDeref(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	deref := funcs["deref"].(func(*string) string)

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, deref(nil))
	})
	t.Run("non-nil", func(t *testing.T) {
		t.Parallel()
		s := "hello"
		assert.Equal(t, "hello", deref(&s))
	})
}

func TestDerefInt(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	derefInt := funcs["deref_int"].(func(*int) int)

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		assert.Zero(t, derefInt(nil))
	})
	t.Run("non-nil", func(t *testing.T) {
		t.Parallel()
		i := 42
		assert.Equal(t, 42, derefInt(&i))
	})
}

func TestDerefStrings(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	derefStrings := funcs["deref_strings"].(func(*[]string) []string)

	t.Run("nil", func(t *testing.T) {
		t.Parallel()
		assert.Nil(t, derefStrings(nil))
	})
	t.Run("non-nil", func(t *testing.T) {
		t.Parallel()
		s := []string{"a", "b"}
		assert.Equal(t, []string{"a", "b"}, derefStrings(&s))
	})
}

func TestNot(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	not := funcs["not"].(func(bool) bool)

	trueResult := not(true)
	falseResult := not(false)
	assert.False(t, trueResult)
	assert.True(t, falseResult)
}

func TestJoinStrings(t *testing.T) {
	t.Parallel()
	funcs := templateFuncs()
	joinStrings := funcs["join_strings"].(func([]string, string) string)

	joined := joinStrings([]string{"a", "b", "c"}, ", ")
	assert.Equal(t, "a, b, c", joined)
}
