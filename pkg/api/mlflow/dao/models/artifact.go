package models

import (
	"time"

	"github.com/google/uuid"
)

// Artifact represents the artifact model.
type Artifact struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Iter      int64     `gorm:"index"`
	Step      int64     `gorm:"default:0;not null"`
	Run       Run
	RunID     string `gorm:"column:run_uuid;not null;index;constraint:OnDelete:CASCADE"`
	Index     int64
	Width     int64
	Height    int64
	Format    string
	Caption   string
	BlobURI   string
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
