package models

import "gorm.io/gorm"

type Friendship struct {
	gorm.Model
	UserID   uint   `json:"user_id"` // Yang me-request
	FriendID uint   `json:"friend_id"` // Yang di-request
	Status   string `json:"status" gorm:"default:'pending'"` // 'pending' atau 'accepted'
	
	// Relasi untuk mengambil data user
	Friend   User   `json:"friend" gorm:"foreignKey:FriendID"`
	User     User   `json:"requester" gorm:"foreignKey:UserID"`
}