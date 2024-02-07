package models

import (
	"context"
	"crypto/sha256"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// DefaultContext is the default metric context
var DefaultContext = Context{Json: JSONB("{}")}

// Metric represents model to work with `metrics` table.
type Metric struct {
	Key       string  `gorm:"type:varchar(250);not null;primaryKey"`
	Value     float64 `gorm:"type:double precision;not null;primaryKey"`
	Timestamp int64   `gorm:"not null;primaryKey"`
	RunID     string  `gorm:"column:run_uuid;not null;primaryKey;index"`
	Step      int64   `gorm:"default:0;not null;primaryKey"`
	IsNan     bool    `gorm:"default:false;not null;primaryKey"`
	Iter      int64   `gorm:"index"`
	ContextID uint    `gorm:"not null;primaryKey"`
	Context   Context
}

// UniqueKey is a compound unique key for this metric series.
func (m Metric) UniqueKey() string {
	return fmt.Sprintf("%v-%v-%v", m.RunID, m.Key, m.ContextID)
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
	ContextID uint `gorm:"not null;primaryKey"`
	Context   Context
}

// UniqueKey is a compound unique key for this metric series.
func (m LatestMetric) UniqueKey() string {
	return fmt.Sprintf("%v-%v-%v", m.RunID, m.Key, m.ContextID)
}

// JSONB defined JSONB data type, need to implements driver.Valuer, sql.Scanner interface
type JSONB json.RawMessage

// Value return json value, implement driver.Valuer interface
func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// Scan scan value into Jsonb, implements sql.Scanner interface
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = JSONB("null")
		return nil
	}
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			bytes = make([]byte, len(v))
			copy(bytes, v)
		}
	case string:
		bytes = []byte(v)
	default:
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	result := json.RawMessage(bytes)
	*j = JSONB(result)
	return nil
}

// MarshalJSON to output non base64 encoded []byte
func (j JSONB) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON to deserialize []byte
func (j *JSONB) UnmarshalJSON(b []byte) error {
	result := json.RawMessage{}
	err := result.UnmarshalJSON(b)
	*j = JSONB(result)
	return err
}

func (j JSONB) String() string {
	return string(j)
}

// GormDataType gorm common data type
func (JSONB) GormDataType() string {
	return "json"
}

// GormDBDataType gorm db data type
func (JSONB) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	return "JSONB"
}

func (js JSONB) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if len(js) == 0 {
		return gorm.Expr("NULL")
	}

	data, _ := js.MarshalJSON()
	return gorm.Expr("?", string(data))
}

// Context represents model to work with `contexts` table.
type Context struct {
	ID   uint  `gorm:"primaryKey;autoIncrement"`
	Json JSONB `gorm:"not null;unique;index"`
}

// GetJsonHash returns hash of the Context.Json
func (c Context) GetJsonHash() string {
	hash := sha256.Sum256(c.Json)
	return string(hash[:])
}
