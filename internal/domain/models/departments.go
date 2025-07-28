package models

type Department struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Section struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parent_id,omitempty"` // Optional, can be used to specify parent section
}
