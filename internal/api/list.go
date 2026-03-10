package api

import "encoding/base64"

const MaxLimit = 64

// obfuscateCursor encodes a string into an opaque cursor.
func obfuscateCursor(name string) Cursor {
	return base64.URLEncoding.EncodeToString([]byte(name))
}

// deobfuscateCursor decodes an opaque cursor string back into a string.
func deobfuscateCursor(cursor Cursor) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// paginate determines if there are more results than the requested limit and generates
// a cursor for the next page if needed. The getCursorValue parameter is a function that extracts
// the cursor value from the last item within the limit. It returns true if there are more results
// available, along with an encoded cursor pointer for fetching the next page (or nil if no more results).
func paginate(count int, limit int, makeDeobfuscatedCursor func() string) (more bool, nextCursor *string) {
	if count > limit {
		encoded := obfuscateCursor(makeDeobfuscatedCursor())
		return true, &encoded
	}
	return false, nil
}

// parseLimit returns a validated limit value, defaulting to and capping at MaxLimit.
func parseLimit(params *int) int {
	limit := MaxLimit
	if params != nil && *params > 0 {
		limit = *params
	}
	return min(limit, MaxLimit)
}
