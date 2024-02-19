package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

// Dashboard represents the dashboard model.
type Dashboard struct {
	Base
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AppID       *uuid.UUID `gorm:"type:uuid" json:"app_id"`
	App         App        `json:"-"`
}

// MarshalJSON marshals the dashboard model to json.
func (d Dashboard) MarshalJSON() ([]byte, error) {
	type jsonDashboard struct {
		Dashboard
		AppType *string `json:"app_type"`
	}
	jd := jsonDashboard{
		Dashboard: d,
	}
	if d.App.IsArchived {
		jd.AppID = nil
	} else {
		jd.AppType = &d.App.Type
	}
	return json.Marshal(jd)
}
