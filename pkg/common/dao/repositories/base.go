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

// NewBaseRepository creates new Base repository.
func NewBaseRepository(db *gorm.DB) *BaseRepository {
	return &BaseRepository{db: db}
}

// GetDB returns current DB instance.
func (r BaseRepository) GetDB() *gorm.DB {
	return r.db
}
