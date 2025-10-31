package api

import "encoding/base64"

const MaxLimit = 64

// encodeCursor encodes a string into an opaque cursor
func encodeCursor(name string) Cursor {
	return base64.URLEncoding.EncodeToString([]byte(name))
}

// decodeCursor decodes an opaque cursor string back into a tag category name
func decodeCursor(cursor Cursor) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(cursor)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}
