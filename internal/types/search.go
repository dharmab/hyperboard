package types

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

type PostSearch struct {
	Tags        []TagName
	ExcludeTags []TagName
	Sort        string
	Tagged      TaggedFilter
	TypeImage   bool // Filter to image posts (type:image)
	TypeVideo   bool // Filter to video posts (type:video)
	TypeAudio   bool // Filter to posts with audio (type:audio)
}
