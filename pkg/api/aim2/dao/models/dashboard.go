package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Dashboard struct {
	Base
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AppID       *uuid.UUID `gorm:"type:uuid" json:"app_id"`
	App         App        `json:"-"`
}

func (d Dashboard) MarshalJSON() ([]byte, error) {
	type localDashboard Dashboard
	type jsonDashboard struct {
		localDashboard
		AppType *string `json:"app_type"`
	}
	jd := jsonDashboard{
		localDashboard: localDashboard(d),
	}
	if d.App.IsArchived {
		jd.AppID = nil
	} else {
		jd.AppType = &d.App.Type
	}
	return json.Marshal(jd)
}
