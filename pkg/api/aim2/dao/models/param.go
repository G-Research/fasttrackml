package models

import (
	"strconv"
	"strings"
)

// Param represents model to work with `params` table.
type Param struct {
	Key        string   `gorm:"type:varchar(250);not null;primaryKey"`
	Value      string   `gorm:"type:varchar(500);not null"`
	RunID      string   `gorm:"column:run_uuid;not null;primaryKey;index"`
}

// ValueTyped returns value held by this Param as any but with correct underlying type.
func (p Param) ValueTyped() any {
	if strings.Contains(p.Value, ".") {
		floatVal, err := strconv.ParseFloat(p.Value, 64)
		if err == nil {
			return floatVal
		}
	}
	intVal, err := strconv.ParseInt(p.Value, 10, 64)
	if err == nil {
		return intVal
	}
	return p.Value
}
