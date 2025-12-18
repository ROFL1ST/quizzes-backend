package models

import "gorm.io/gorm"

type Item struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Type        string `json:"type"`
	AssetURL    string `json:"asset_url"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
}

type UserItem struct {
	UserID     uint `json:"user_id" gorm:"primaryKey"`
	ItemID     uint `json:"item_id" gorm:"primaryKey"`
	IsEquipped bool `json:"is_equipped" gorm:"default:false"` // Sedang dipakai atau tidak
	Item       Item `json:"item" gorm:"foreignKey:ItemID"`
}
