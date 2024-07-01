package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Base represents a base model which other models just inherited.
type Base struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsArchived bool      `json:"-"`
}

// App represents a model to work with `apps` table.
type App struct {
	Base
	Type        string    `gorm:"not null" json:"type"`
	State       AppState  `json:"state"`
	Namespace   Namespace `json:"-"`
	NamespaceID uint      `gorm:"not null" json:"-"`
}

// AppState represents the state of App entity.
type AppState map[string]any

// Value implements Gorm interface.
func (s AppState) Value() (driver.Value, error) {
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

// Scan implements Gorm interface.
func (s *AppState) Scan(v interface{}) error {
	var nullS sql.NullString
	if err := nullS.Scan(v); err != nil {
		return err
	}
	if nullS.Valid {
		return json.Unmarshal([]byte(nullS.String), s)
	}
	return nil
}

// GormDataType implements Gorm interface.
func (s AppState) GormDataType() string {
	return "text"
}
