package search

const (
	// TagImage is an automatic tag applied to all image posts.
	TagImage = "type:image"
	// TagVideo is an automatic tag applied to all video posts.
	TagVideo = "type:video"
	// TagAudio is an automatic tag applied to all video or audio posts which contain audio.
	TagAudio = "type:audio"
)

const (
	// FilterTagged is a boolean filter that matches posts which have tags other than automatic tags.
	// tagged:true matches posts with at least one non-automatic tag.
	// tagged:false matches posts with no non-automatic tags (useful for finding untagged posts).
	FilterTagged = "tagged:"
)

const (
	// SortRandom is a shuffled ordering. It is not randomized every query, but rather is a pre-shuffled ordering which is periodically reshuffled.
	SortRandom = "random"
	// SortCreatedAt orders posts by their creation..
	SortCreatedAt = "created"
	// SortUpdatedAt orders posts by their last update.
	SortUpdatedAt = "updated"
)
