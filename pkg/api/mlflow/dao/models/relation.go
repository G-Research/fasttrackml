package models

import (
	"github.com/google/uuid"
)

// RoleNamespace represents a model to work with `role_relations` table.
// Model holds relations between Role and Namespace models.
type RoleNamespace struct {
	Base
	Role        Role      `gorm:"constraint:OnDelete:CASCADE"`
	RoleID      uuid.UUID `gorm:"not null;index:,unique,composite:relation"`
	Namespace   Namespace `gorm:"constraint:OnDelete:CASCADE"`
	NamespaceID uint      `gorm:"not null;index:,unique,composite:relation"`
}
