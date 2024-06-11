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

// SharedTag represents model to work with `shared_tags` for Aim UI.
type SharedTag struct {
	ID          uuid.UUID `gorm:"column:id;not null"`
	IsArchived  bool      `gorm:"not null,default:false"`
	Key         string    `gorm:"type:varchar(250);not null;primaryKey"`
	Color       string    `gorm:"type:varchar(7);null"`
	Description string    `gorm:"type:varchar(500);null`
	NamespaceID uint      `gorm:"not null;primaryKey"`
	Runs        []Run     `gorm:"many2many:run_tag_datas"`
}

// BeforeCreate supplies a UUID for SharedTag.
func (sharedTag *SharedTag) BeforeCreate(tx *gorm.DB) error {
	sharedTag.ID = uuid.New()
	return nil
}
