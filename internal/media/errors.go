package media

import "errors"

// ErrInvalidMedia indicates that submitted media cannot be decoded or processed.
var ErrInvalidMedia = errors.New("invalid media")
