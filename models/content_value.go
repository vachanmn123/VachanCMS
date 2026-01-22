package models

// This is the data/<CTSlug>/config.json file that will be used by the CMS to keep track of content values, the map approach is to make it faster to look up filenames by content value ID
type ContentValueConfigFile struct {
	TotalPages   int               `json:"total_pages"`
	TotalItems   int               `json:"total_items"`
	ItemsPerPage int               `json:"items_per_page"`
	Items        map[string]int    `json:"items"` // map of content value ID to page
	Slugs        map[string]string `json:"slugs"` // map of slug to content value ID
}

// This is the data/<CTSlug>/index-<PAGE>.json file that will be used by the CMS to list content values
type ContentValueIndexFile struct {
	Page  int            `json:"page"`
	Items []ContentValue `json:"items"`
}

type ContentValue struct {
	Id    string         `json:"id,omitempty"`
	Slug  string         `json:"slug,omitempty"`
	Value map[string]any `json:"values" binding:"required"`
}
