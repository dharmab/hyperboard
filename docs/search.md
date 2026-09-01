# Search

A query is a comma-separated list of terms. Whitespace surrounding each term is trimmed, and empty terms are ignored.

```text
national park, -utah, type:image, sort:created, order:desc
```

Terms and operators are case-sensitive. Tag and alias matching is also case-sensitive. All filters are combined with **AND**:

- every included tag must match;
- no excluded tag may match; and
- every content type, tag-presence, and date condition must match.

There is no comma-escaping syntax. Tag names and aliases therefore cannot contain commas, start with `-`, or start with a reserved operator prefix: `sort:`, `order:`, `tagged:`, `type:`, `created_after:`, or `created_before:`.

### Tag Filters

| Syntax | Effect |
|--------|--------|
| `zion national park` | Include posts tagged `zion national park`, directly or through a cascade |
| `-utah` | Exclude posts tagged `utah`, directly or through a cascade |
| `national park, utah` | Require both `national park` AND `utah` |
| `national park, -utah` | Require `national park` and reject `utah` |

Tag aliases are resolved for both inclusion and exclusion. For example, if `zion` is an alias of `zion national park`, the two names produce the same filter.

Tag cascades are one-directional and one level deep. If A cascades to B, a post directly tagged A matches a search for B. If B also cascades to C, a post directly tagged only A does not match C.

### Content Type Filters

| Syntax | Effect |
|--------|--------|
| `type:image` | Require an image MIME type (`image/*`) |
| `type:video` | Require a video MIME type (`video/*`) |
| `type:audio` | Require the post's audio flag |
| `-type:image` | Reject image MIME types |
| `-type:video` | Reject video MIME types |
| `-type:audio` | Require the post not to have audio |

Type filters combine with all other filters using AND. `type:audio` is independent of MIME type, so media such as an animated image with audio can match `type:image,type:audio`. Mutually exclusive filters, such as `type:image,type:video` or `type:image,-type:image`, return no posts.

### Tag Presence Filter

| Syntax | Effect |
|--------|--------|
| `tagged:true` | Require at least one directly assigned tag |
| `tagged:false` | Require no directly assigned tags |

Cascade-derived tags do not count as directly assigned tags for this filter. This makes `tagged:false` useful for finding posts that still need explicit tagging.

### Date Range Filters

| Syntax | Effect |
|--------|--------|
| `created_after:2025-01-01T00:00:00Z` | Require a creation time strictly after the timestamp |
| `created_before:2025-06-01T00:00:00Z` | Require a creation time strictly before the timestamp |

Timestamps must use RFC 3339 format; fractional seconds are accepted. Both bounds are exclusive, so a post created exactly at either timestamp does not match. Both filters can be combined to form an open date range.

### Sort Options

| Syntax | Effect |
|--------|--------|
| `sort:created` | Sort by creation time (the API default) |
| `sort:updated` | Sort by last update time |
| `sort:file-size` | Sort by content file size |
| `sort:random` | Use a deterministic shuffle that changes every six hours |

Random order is stable within a rotating six-hour window. `order:asc` and `order:desc` do not affect random sorting. If pagination continues after the random window changes, it restarts in the new ordering and may repeat posts.

### Sort Direction

| Syntax | Effect |
|--------|--------|
| `order:desc` | Descending (default): newest or largest first |
| `order:asc` | Ascending: oldest or smallest first |

### Examples

| Query | Result |
|-------|--------|
| `national park, utah, -zion national park` | Require `national park` and `utah`, excluding `zion national park` |
| `sort:random, type:video` | Videos in the current six-hour random order |
| `tagged:false, type:image` | Untagged images |
| `created_after:2025-01-01T00:00:00Z, yosemite national park` | Posts tagged `yosemite national park` created strictly after Jan 1, 2025 |
| `sort:created, order:asc` | All posts, oldest first |
| `sort:file-size, order:desc, -type:video` | Non-video posts, largest first |
| `type:video, -type:audio` | Videos without audio |

## Tags

Tags are labels assigned to posts. Each tag has a unique name, an optional description, an optional category, and zero or more aliases and cascading tags.

### Aliases

A tag can have one or more aliases (alternate names). For example, if `zion national park` has the alias `zion`, searching for either name returns the same results.

Tags can be merged by converting one tag into an alias of another. The merge retags posts from the source to the target and makes the source name an alias of the target.

### Cascading Tags

When tag A cascades to tag B, posts directly tagged A also appear in searches for B without B being explicitly assigned. Cascades are one-directional and are not transitive beyond one level.

For example, `zion national park` could cascade to `utah` and `national park`. A post tagged `zion national park` then matches searches for either cascaded tag. A post tagged only `utah` does not match `zion national park`.

Cascade changes apply immediately to existing posts because the relationship is evaluated during search. A tag's post count includes directly tagged posts and posts that reach it through one cascade.

### Tag Categories

Tags can optionally belong to a category. Categories are organizational groups that give their tags a customizable badge color. A tag inherits its color from its category; uncategorized tags use the default color. Categories do not change post-search matching.
