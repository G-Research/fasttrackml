package fixtures

import (
	"gorm.io/gorm"
)

// ProjectFixtures represents data fixtures object.
type ProjectFixtures struct {
	baseFixtures
}

// NewProjectFixtures creates new instance of ProjectFixtures.
func NewProjectFixtures(db *gorm.DB) (*ProjectFixtures, error) {
	return &ProjectFixtures{
		baseFixtures: baseFixtures{db: db},
	}, nil
}
