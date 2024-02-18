package models

import "github.com/google/uuid"

type Dashboard struct {
	Base
	Name        string     `json:"name"`
	Description string     `json:"description"`
	AppID       *uuid.UUID `gorm:"type:uuid" json:"app_id"`
	App         App        `json:"-"`
}
