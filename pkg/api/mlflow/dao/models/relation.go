package models

import (
	"time"

	"github.com/google/uuid"
)

// RoleNamespace represents model to work with `role_relations` table.
// Model holds relations between Role and Namespace models.
type RoleNamespace struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	Role        Role      `gorm:"constraint:OnDelete:CASCADE"`
	RoleID      uuid.UUID `gorm:"not null;index:,unique,composite:relation"`
	Namespace   Namespace `gorm:"constraint:OnDelete:CASCADE"`
	NamespaceID uuid.UUID `gorm:"not null;index:,unique,composite:relation"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
