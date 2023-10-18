package models

// Metric represents model to work with `metrics` table.
type Metric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null;primaryKey"`
	Timestamp int64   `gorm:"not null;primaryKey"`
	RunID     string  `gorm:"column:run_uuid;not null;primaryKey;index"`
	Step      int64   `gorm:"default:0;not null;primaryKey"`
	IsNan     bool    `gorm:"default:false;not null;primaryKey"`
	Iter      int64   `gorm:"index"`
	Context   string  `gorm:"type:varchar(250);not null;primaryKey"`
}

// LatestMetric represents model to work with `last_metrics` table.
type LatestMetric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null"`
	Timestamp int64
	Step      int64  `gorm:"not null"`
	IsNan     bool   `gorm:"not null"`
	RunID     string `gorm:"column:run_uuid;not null;primaryKey;index"`
	LastIter  int64
	Context   string `gorm:"type:varchar(250);not null;primaryKey"`
}
