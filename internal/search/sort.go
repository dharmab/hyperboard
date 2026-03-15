package search

// Sort represents a sort ordering for search results.
type Sort string

const (
	// SortRandom is a shuffled ordering. It is not randomized every query, but rather is a pre-shuffled ordering which is periodically reshuffled.
	SortRandom Sort = "sort:random"
	// SortCreatedAt orders posts by their creation.
	SortCreatedAt Sort = "sort:created"
	// SortUpdatedAt orders posts by their last update.
	SortUpdatedAt Sort = "sort:updated"
)

// Order represents the sort direction for search results.
type Order string

const (
	// OrderAsc sorts results in ascending order.
	OrderAsc Order = "order:asc"
	// OrderDesc sorts results in descending order (default).
	OrderDesc Order = "order:desc"
)
