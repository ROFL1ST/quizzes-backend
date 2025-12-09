package models

import "gorm.io/gorm"

type Friendship struct {
	gorm.Model
	UserID   uint `json:"user_id"`
	FriendID uint `json:"friend_id"`
	Friend   User `json:"friend" gorm:"foreignKey:FriendID"`
}