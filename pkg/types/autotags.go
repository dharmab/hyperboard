package types

const (
	// TagImage is an automatic tag applied to all image posts.
	TagImage = "type:image"
	// TagVideo is an automatic tag applied to all video posts.
	TagVideo = "type:video"
	// TagAudio is an automatic tag applied to all video or audio posts which contain audio.
	TagAudio = "type:audio"
)

const (
	// FilterIsFavorite is a boolean filter that matches posts which are favorited.
	FilterIsFavorite = "favorite:"
	// FilterTagged is a boolean filter that matches posts which have tags other than automatic tags.
	// It is mostly useful when inverted to find posts which need to be tagged.
	FilterTagged = "tagged:"
	// FilterCreatedAt is a time or duration filter that matches posts created before or after a specific time, or within a specific duration.
	FilterCreatedAt = "created:"
	// FilterUpdatedAt is a time or duration filter that matches posts updated before or after a specific time, or within a specific duration.
	FilterUpdatedAt = "updated:"
)

const (
	// SortRandom is a shuffled ordering. It is not randomized every query, but rather is a pre-shuffled ordering which is periodically reshuffled.
	SortShuffle = "shuffle"
	// SortCreatedAt orders posts by their creation..
	SortCreatedAt = "created"
	// SortUpdatedAt orders posts by their last update.
	SortUpdatedAt = "updated"
	// SortFileSize orders posts by their file size on disk.
	SortFileSize = "file-size"
	// SortResolution orders posts by their resolution (pixel width x height).
	SortResolution = "resolution"
	// SortDuration sorts video and audio posts by their runtime. Images have a duration of 0.
	SortDuration = "duration"
)
