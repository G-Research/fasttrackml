package repositories

import "gorm.io/gorm"

// BaseRepositoryProvider provides base repository interface.
type BaseRepositoryProvider interface {
	// GetDB returns current DB instance.
	GetDB() *gorm.DB
}

// BaseRepository represents base repository object.
type BaseRepository struct {
	db *gorm.DB
}

// GetDB returns current DB instance.
func (r BaseRepository) GetDB() *gorm.DB {
	return r.db
}
