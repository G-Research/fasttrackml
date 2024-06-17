package models

// Log represents a row of the `logs` table.
type Log struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	Value     string `gorm:"not null"`
	RunID     string `gorm:"column:run_uuid;not null;index"`
	Timestamp int64  `gorm:"not null;index"`
}
