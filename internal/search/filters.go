package search

// TaggedFilter represents the state of the tagged: filter.
type TaggedFilter int

const (
	// TaggedFilterNone means no tagged: filter is applied.
	TaggedFilterNone TaggedFilter = iota
	// TaggedFilterTrue matches posts with at least one non-automatic tag.
	TaggedFilterTrue
	// TaggedFilterFalse matches posts with no non-automatic tags.
	TaggedFilterFalse
)
