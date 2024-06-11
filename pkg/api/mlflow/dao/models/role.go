package models

// Role represents model to work with `roles` table.
type Role struct {
	Base
	Name string `gorm:"unique;index;not null"`
}
