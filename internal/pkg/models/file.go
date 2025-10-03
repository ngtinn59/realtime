package models

import (
	"time"

	"gorm.io/gorm"
)

// File represents an uploaded file
type File struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UploaderID uint           `gorm:"not null;index" json:"uploader_id"`
	Uploader   User           `gorm:"foreignKey:UploaderID" json:"uploader,omitempty"`
	Filename   string         `gorm:"not null;size:255" json:"filename"`
	OriginalName string       `gorm:"not null;size:255" json:"original_name"`
	MimeType   string         `gorm:"size:100" json:"mime_type"`
	Size       int64          `gorm:"not null" json:"size"` // in bytes
	URL        string         `gorm:"not null;size:500" json:"url"`
	Path       string         `gorm:"not null;size:500" json:"path"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name
func (File) TableName() string {
	return "files"
}
