package models

// Log represents a row of the `logs` table.
type Log struct {
	Timestamp int64  `gorm:"not null;primaryKey"`
	Value     string `gorm:"type:varchar(5000)"`
	RunID     string `gorm:"column:run_uuid;not null;primaryKey;index"`
}
