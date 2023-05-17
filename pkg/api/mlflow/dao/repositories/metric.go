package repositories

import (
	"gorm.io/gorm"
)

// MetricRepositoryProvider provides an interface to work with models.Metric entity.
type MetricRepositoryProvider interface {
	BaseRepositoryProvider
}

// MetricRepository repository to work with models.Metric entity.
type MetricRepository struct {
	BaseRepository
}

// NewMetricRepository creates repository to work with models.Metric entity.
func NewMetricRepository(db *gorm.DB) *MetricRepository {
	return &MetricRepository{
		BaseRepository{
			db: db,
		},
	}
}
