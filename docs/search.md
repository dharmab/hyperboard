# Search

This document explains how tags and searching by tags works in Hyperboard.

## Tags

Tags are labels assigned to posts. Each tag has a unique name, an optional description, an optional category, and zero or more aliases and cascading tags.

### Aliases

A tag can have one or more aliases (alternate names). For example, if the tag `zion national park` has an alias `zion`, searching for `zion` returns the same results as searching for `zion national park`.

Tags can be merged by converting one tag into an alias of another. When merged, all posts with the source tag are re-tagged with the target tag, and the source tag's name becomes an alias of the target.

### Cascading Tags

Tags can cascade to other tags. When tag A cascades to tag B, all posts tagged with A also appear in searches for B. Cascades are one-directional — A cascading to B does not mean B cascades to A.

For example, the tag `zion national park` could cascade to `utah` and `national park`. A post tagged `zion national park` would then appear in searches for `utah` and `national park` without needing to be explicitly tagged with those terms. However, searching for `zion national park` would not return posts that are only tagged `utah`.

Changes to a tag's cascading tags automatically apply to all existing posts with that tag.

A tag's post count includes both directly tagged posts and posts that cascade to it.

### Tag Categories

Tags can optionally belong to a tag category. Categories are organizational groups that give their tags a customizable badge color. A tag inherits its color from its category. If a tag has no category, it uses a default color.

## Search Syntax

Searches are entered as comma-separated terms. Whitespace around each term is trimmed.

### Tag Filters

| Syntax | Effect |
|--------|--------|
| `zion national park` | Include posts tagged `zion national park` (directly or via cascade) |
| `-utah` | Exclude posts tagged `utah` |
| `national park, utah` | Posts tagged with both `national park` AND `utah` |
| `national park, -utah` | Posts tagged `national park` but not `utah` |

### Content Type Filters

| Syntax | Effect |
|--------|--------|
| `type:image` | Posts with an image MIME type |
| `type:video` | Posts with a video MIME type |
| `type:audio` | Posts with audio tracks |

Type filters can be combined with tag filters and with each other. Prefix with `-` to exclude (e.g. `-type:audio`).

### Tag Presence Filter

| Syntax | Effect |
|--------|--------|
| `tagged:true` | Posts with at least one tag |
| `tagged:false` | Posts with no tags |

This is handy for finding posts that need tagging.

### Date Range Filters

| Syntax | Effect |
|--------|--------|
| `created_after:2025-01-01T00:00:00Z` | Posts created after the given timestamp |
| `created_before:2025-06-01T00:00:00Z` | Posts created before the given timestamp |

Timestamps must be in RFC 3339 format. Both filters can be used together to define a range.

### Sort Options

| Syntax | Effect |
|--------|--------|
| `sort:created` | Sort by creation time (default) |
| `sort:updated` | Sort by last update time |
| `sort:random` | Random shuffle (new ordering every ~6 hours) |

### Sort Direction

| Syntax | Effect |
|--------|--------|
| `order:desc` | Descending / newest first (default) |
| `order:asc` | Ascending / oldest first |

### Examples

| Query | Result |
|-------|--------|
| `national park, utah, -zion national park` | Posts tagged `national park` and `utah`, excluding `zion national park` |
| `sort:random, type:video` | Random video shuffle |
| `tagged:false, type:image` | Untagged images |
| `created_after:2025-01-01T00:00:00Z, yosemite national park` | Posts tagged `yosemite national park` created after Jan 1, 2025 |
| `sort:created, order:asc` | All posts, oldest first |
