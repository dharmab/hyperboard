# Tag Filters

Tag filters are configurable buttons displayed on the posts page that let you quickly add or cycle through search tags with a single click.

## Configuration

Tag filters are configured via the `--tag-filters` flag or the `HYPERBOARD_WEB_TAG_FILTERS` environment variable. The value is a JSON array of filter objects.

Each filter object has:

| Field | Type | Description |
|-------|------|-------------|
| `label` | string | Button label text (or a material icon reference) |
| `tags` | string[] | One or more tag search terms to cycle through |

Example:

```
--tag-filters='[{"label":"Favorites","tags":["favorite"]},{"label":"Genre","tags":["portrait","landscape","street"]},{"label":"No People","tags":["-person"]}]'
```

Or as an environment variable:

```
HYPERBOARD_WEB_TAG_FILTERS='[{"label":"Favorites","tags":["favorite"]},{"label":"Genre","tags":["portrait","landscape","street"]}]'
```

## How Tag Filter Buttons Work

Each button cycles through its list of tags:

1. **First click**: Adds the first tag in the list to the search.
2. **Subsequent clicks**: Replaces the current tag with the next one in the list.
3. **After the last tag**: Removes the tag from the search (cycles back to no filter).

For example, a filter with `"tags": ["portrait", "landscape", "street"]` cycles through four states: `portrait` -> `landscape` -> `street` -> all photos.

Tags prefixed with `-` exclude instead of include. A filter with `"tags": ["black_and_white", "-black_and_white"]` cycles through three states: only B&W photos -> no B&W photos -> all photos.

Active buttons are highlighted to indicate which tag is currently applied.

## Material Icon Labels

Instead of text labels, you can use [Material Symbols](https://fonts.google.com/icons) as button icons. Prefix the label with `material-icons-` followed by the icon name.

For example, to use the `star` icon for favorites:

```json
{"label": "material-icons-star", "tags": ["favorite"]}
```

To use the `landscape` icon for outdoor shots:

```json
{"label": "material-icons-landscape", "tags": ["landscape", "nature", "cityscape"]}
```

Browse available icons at https://fonts.google.com/icons. Use the icon name exactly as shown (lowercase, underscores for spaces).

## Quick Tag

The `--quick-tag` flag (or `HYPERBOARD_WEB_QUICK_TAG` environment variable) configures a single tag that can be toggled on individual posts using the `f` keyboard shortcut. Posts with the quick tag display a star indicator next to their file size.

Example:

```
--quick-tag=favorite
```

## Full Example

```
hyperboard-web \
  --tag-filters='[{"label":"material-icons-star","tags":["favorite"]},{"label":"Genre","tags":["portrait","landscape","street","macro"]},{"label":"B&W","tags":["black_and_white"]}]' \
  --quick-tag=favorite
```
