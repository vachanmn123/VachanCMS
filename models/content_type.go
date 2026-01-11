package models

type ContentTypeField struct {
	FieldName string `json:"field_name" binding:"required"`
	// text, number, boolean, select, multiselect, date, etc.
	FieldType  string   `json:"field_type" binding:"required"`
	IsRequired bool     `json:"is_required"`
	Options    []string `json:"options,omitempty"`
}

type ContentType struct {
	Id     string             `json:"id,omitempty"`
	Name   string             `json:"name" binding:"required"`
	Slug   string             `json:"slug" binding:"required"`
	Fields []ContentTypeField `json:"fields" binding:"required"`
}
