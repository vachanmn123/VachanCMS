package models

// This is the media/config.json that contains page data and is a directory
type MediaConfigFile struct {
	TotalPages   int            `json:"total_pages"`
	TotalItems   int            `json:"total_items"`
	ItemsPerPage int            `json:"items_per_page"`
	Items        map[string]int `json:"items"` // map of media ID to page number
}

// This is the media/index-<PAGE>.json that will hold the list of all media files in the repo.
type MediaIndexFile struct {
	Page  int         `json:"page"`
	Media []MediaFile `json:"media"`
}

type MediaFile struct {
	Id       string `json:"id"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
}
