package api

import "unicode"

// isValidName reports whether name begins with a unicode letter or digit,
// does not end with whitespace, and does not contain consecutive whitespace.
func isValidName(name string) bool {
	var prev rune
	for i, r := range name {
		if i == 0 && !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return false
		}
		if unicode.IsSpace(r) && unicode.IsSpace(prev) {
			return false
		}
		prev = r
	}
	return prev != 0 && !unicode.IsSpace(prev)
}

// isValidHexColor reports whether s is a valid 6-digit hex color (e.g. "#ff0000").
func isValidHexColor(s string) bool {
	if len(s) != 7 || s[0] != '#' {
		return false
	}
	for _, c := range s[1:] {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') && (c < 'A' || c > 'F') {
			return false
		}
	}
	return true
}
