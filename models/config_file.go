package models

type ConfigFile struct {
	SiteName           string        `json:"site_name" binding:"required"`
	ContentTypes       []ContentType `json:"content_types" binding:"required"`
	InitializationDate string        `json:"initialization_date" binding:"required"`
}
