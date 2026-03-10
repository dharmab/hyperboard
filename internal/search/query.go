package search

import "github.com/dharmab/hyperboard/pkg/types"

type Query struct {
	IncludedTags []types.TagName
	ExcludedTags []types.TagName
	Sort         string
	TaggedTrue   bool // Filter to posts with at least one non-automatic tag (tagged:true)
	TaggedFalse  bool // Filter to posts with no non-automatic tags (tagged:false)
	TypeImage    bool // Filter to image posts (type:image)
	TypeVideo    bool // Filter to video posts (type:video)
	TypeAudio    bool // Filter to posts with audio (type:audio)
}
