package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Translation struct {
	gorm.Model
	Section      string         `gorm:"index;not null" json:"section"`           // e.g. "auth", "navbar"
	Key          string         `gorm:"index;not null" json:"key"`               // e.g. "loginTitle"
	Translations datatypes.JSON `gorm:"type:jsonb;not null" json:"translations"` // {"id": "...", "en": "..."}
}

// TableName overrides the table name used by Translation to `app_translations`
func (Translation) TableName() string {
	return "app_translations"
}
