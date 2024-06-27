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

// SharedTag represents model to work with `shared_tags` (Aim UI tag).
type SharedTag struct {
	ID          uuid.UUID `gorm:"column:id;not null;primaryKey"`
	IsArchived  bool      `gorm:"not null,default:false"`
	Name        string    `gorm:"type:varchar(250);not null"`
	Color       string    `gorm:"type:varchar(7);null"`
	Description string    `gorm:"type:varchar(500);null"`
	NamespaceID uint      `gorm:"not null"`
	Runs        []Run     `gorm:"many2many:run_shared_tags"`
}

// BeforeCreate supplies a UUID for SharedTag.
func (sharedTag *SharedTag) BeforeCreate(tx *gorm.DB) error {
	sharedTag.ID = uuid.New()
	return nil
}
