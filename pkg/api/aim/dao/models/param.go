package models

import (
	"fmt"
)

// Param represents model to work with `params` table.
type Param struct {
	Key        string   `gorm:"type:varchar(250);not null;primaryKey"`
	ValueStr   *string  `gorm:"type:varchar(500)"`
	ValueInt   *int64   `gorm:"type:bigint"`
	ValueFloat *float64 `gorm:"type:float"`
	RunID      string   `gorm:"column:run_uuid;not null;primaryKey;index"`
}

// Value returns the value held by this Param as a string.
func (p Param) ValueString() string {
	switch {
	case p.ValueInt != nil:
		return fmt.Sprintf("%v", *p.ValueInt)
	case p.ValueFloat != nil:
		return fmt.Sprintf("%v", *p.ValueFloat)
	case p.ValueStr != nil:
		return *p.ValueStr
	default:
		return ""
	}
}

// ValueAny returns the value held by this Param as any with underlying type.
func (p Param) ValueAny() any {
	switch {
	case p.ValueInt != nil:
		return *p.ValueInt
	case p.ValueFloat != nil:
		return *p.ValueFloat
	case p.ValueStr != nil:
		return *p.ValueStr
	default:
		return nil
	}
}
