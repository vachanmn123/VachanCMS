package models

type ContentTypeField struct {
	FieldName string `json:"field_name" binding:"required"`
	// Supported field types:
	// - text: string value (single line)
	// - textarea: string value (multi-line, larger text)
	// - number: numeric value (float64)
	// - boolean: true/false value
	// - select: single selection from Options array
	// - media: reference to media file(s) by ID
	//   - If Options contains "multiple", stores array of media IDs
	//   - Otherwise, stores a single media ID string
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
