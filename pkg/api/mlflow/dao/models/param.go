package models

import "fmt"

// Param represents model to work with `params` table.
type Param struct {
	Key        string   `gorm:"type:varchar(250);not null;primaryKey"`
	Value      string   `gorm:"type:varchar(500);not null"`
	ValueInt   *int64   `gorm:"type:bigint;not null"`
	ValueFloat *float64 `gorm:"type:float;not null"`
	RunID      string   `gorm:"column:run_uuid;not null;primaryKey;index"`
}

// ValueString returns the value held by this Param as a string
func (p Param) ValueString() string {
	if p.ValueInt != nil {
		return fmt.Sprintf("%v", p.ValueInt)
	} else if p.ValueFloat != nil {
		return fmt.Sprintf("%v", p.ValueFloat)
	} else {
		return p.Value
	}
}
