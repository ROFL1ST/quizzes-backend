package models

import "gorm.io/gorm"

type Admin struct {
	gorm.Model
	Name     string `json:"name"`
	Username string `json:"username" gorm:"unique;not null"`
	Password string `json:"-"`
	RoleID   uint   `json:"role_id"`
	Role     Role   `json:"role" gorm:"foreignKey:RoleID"`
}