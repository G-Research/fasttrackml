package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/rotisserie/eris"
	"gorm.io/gorm"
)

// Artifact represents the artifact model.
type Artifact struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string    `gorm:"not null;index"`
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

// AfterSave will calculate the iter number for this step sequence based on creation time.
func (u *Artifact) AfterSave(tx *gorm.DB) error {
	if err := tx.Exec(
		`UPDATE artifacts
	         SET iter = rows.new_iter
                 FROM (
                   SELECT id, ROW_NUMBER() OVER (ORDER BY created_at) as new_iter
                   FROM artifacts
                   WHERE run_uuid = ?
                   AND name = ?
                   AND step = ?
                 ) as rows
	         WHERE artifacts.id = rows.id`,
		u.RunID, u.Name, u.Step,
	).Error; err != nil {
		return eris.Wrap(err, "error updating artifacts iter")
	}
	return nil
}
