package types

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

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

// GormValue gorm db actual value
// nolint
func (js JSONB) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if len(js) == 0 {
		return gorm.Expr("NULL")
	}

	data, _ := js.MarshalJSON()
	return gorm.Expr("?", string(data))
}
