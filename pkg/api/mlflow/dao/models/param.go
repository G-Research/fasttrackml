package models

// Param represents model to work with `params` table.
type Param struct {
	Key   string `gorm:"type:varchar(250);not null;primaryKey"`
	Value string `gorm:"type:varchar(500);not null;primaryKey"`
	RunID string `gorm:"column:run_uuid;not null;primaryKey;index"`
}
