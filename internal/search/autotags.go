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
	// TagTaggedTrue is an autotag filter that matches posts with at least one non-automatic tag.
	TagTaggedTrue = "tagged:true"
	// TagTaggedFalse is an autotag filter that matches posts with no non-automatic tags (useful for finding untagged posts).
	TagTaggedFalse = "tagged:false"
)

const (
	// SortRandom is a shuffled ordering. It is not randomized every query, but rather is a pre-shuffled ordering which is periodically reshuffled.
	SortRandom = "random"
	// SortCreatedAt orders posts by their creation..
	SortCreatedAt = "created"
	// SortUpdatedAt orders posts by their last update.
	SortUpdatedAt = "updated"
)
