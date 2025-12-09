package models

import "gorm.io/gorm"

type Role struct {
	gorm.Model
	Name        string  `json:"name" gorm:"unique;not null"`
	Description string  `json:"description"`
	Admins      []Admin `json:"-" gorm:"foreignKey:RoleID"`
}