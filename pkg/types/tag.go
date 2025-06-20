package types

type TagName string

type Tag struct {
	Name     TagName      `json:"name"`
	Category CategoryName `json:"category"`
}

type CategoryName string

type TagCategory struct {
	Name CategoryName `json:"name"`
}
