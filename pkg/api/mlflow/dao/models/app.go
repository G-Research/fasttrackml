package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"

	"time"

	"github.com/google/uuid"
)

type App struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	IsArchived bool      `json:"-"`
	Type  string   `gorm:"not null" json:"type"`
	State AppState `json:"state"`
}

type AppState map[string]any

func (s AppState) Value() (driver.Value, error) {
	v, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(v), nil
}

func (s *AppState) Scan(v interface{}) error {
	var nullS sql.NullString
	if err := nullS.Scan(v); err != nil {
		return err
	}
	if nullS.Valid {
		return json.Unmarshal([]byte(nullS.String), s)
	}
	s = nil
	return nil
}
