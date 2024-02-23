package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Tag represents model to work with `tags` table.
type Tag struct {
	Key   string `gorm:"type:varchar(250);not null;primaryKey"`
	Value string `gorm:"type:varchar(5000)"`
	RunID string `gorm:"column:run_uuid;not null;primaryKey;index"`
}

// TagExtraInfo stores additional tag data for Aim UI.
type TagExtraInfo struct {
	Key         string    `gorm:"type:varchar(250);not null;primaryKey"`
	ID          uuid.UUID `gorm:"column:id;not null"`
	Color       string    `gorm:"type:varchar(7);null"`
	Description string    `gorm:"type:varchar(500);null`
}

// BeforeCreate supplies a UUID for TagExtraInfo.
func (tag *TagExtraInfo) BeforeCreate(tx *gorm.DB) error {
	tag.ID = uuid.New()
	return nil
}
