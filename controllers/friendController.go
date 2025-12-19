package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"time"
	"github.com/gofiber/fiber/v2"
)

type FriendResponse struct {
    ID            uint          `json:"id"`
    Name          string        `json:"name"`
    Username      string        `json:"username"`
    Level         int           `json:"level"`
    EquippedItems []models.Item `json:"equipped_items"` // <--- Tambahkan ini
}

func RequestFriend(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var input struct {
		Username string `json:"username"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var friend models.User
	if err := config.DB.Where("username = ?", input.Username).First(&friend).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	if friend.ID == uint(userID) {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Cannot add yourself", nil)
	}

	var check models.Friendship
	err := config.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friend.ID, friend.ID, userID).First(&check).Error

	if err == nil {
		if check.Status == "accepted" {
			return utils.ErrorResponse(c, fiber.StatusBadRequest, "Already friends", nil)
		}
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Request already sent/pending", nil)
	}

	friendship := models.Friendship{
		UserID:   uint(userID),
		FriendID: friend.ID,
		Status:   "pending",
	}

	if err := config.DB.Create(&friendship).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to send request", err.Error())
	}

	var sentCount int64
	config.DB.Model(&models.Friendship{}).
		Where("user_id = ?", userID).
		Count(&sentCount)

	if sentCount >= 5 {
		utils.UnlockAchievement(uint(userID), 16)
	}

	utils.SendNotification(friend.ID, "info", "Permintaan Teman Baru", "👋 Permintaan teman baru dari "+input.Username, "/friends")
	return utils.SuccessResponse(c, fiber.StatusCreated, "Friend request sent", nil)
}

func ConfirmFriend(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	var input struct {
		RequesterID uint `json:"requester_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	var friendship models.Friendship
	if err := config.DB.Where("user_id = ? AND friend_id = ? AND status = 'pending'", input.RequesterID, myID).First(&friendship).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Friend request not found", nil)
	}

	friendship.Status = "accepted"
	config.DB.Save(&friendship)

	var friendCount int64
	config.DB.Model(&models.Friendship{}).
		Where("(user_id = ? OR friend_id = ?) AND status = 'accepted'", myID, myID).
		Count(&friendCount)

	if friendCount >= 3 {
		utils.UnlockAchievement(uint(myID), 7)
	}
	utils.CheckDailyMissions(uint(myID), "social", 1, "add") // Yang menerima
	utils.CheckDailyMissions(input.RequesterID, "social", 1, "add")   // Yang meminta
	return utils.SuccessResponse(c, fiber.StatusOK, "Friend request accepted", nil)
}

// 3. Refuse Friend (Menolak permintaan)
func RefuseFriend(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	var input struct {
		RequesterID uint `json:"requester_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", err.Error())
	}

	// Hapus request dimana SAYA adalah penerima
	result := config.DB.Where("user_id = ? AND friend_id = ? AND status = 'pending'", input.RequesterID, myID).Delete(&models.Friendship{})

	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Request not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Friend request refused", nil)
}

// 4. Get My Friends (Hanya yang statusnya accepted)
func GetMyFriends(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var friendships []models.Friendship

	// 1. Ambil data friendship (User terlibat & accepted)
	config.DB.Preload("Friend").Preload("User").
		Where("(user_id = ? OR friend_id = ?) AND status = 'accepted'", userID, userID).
		Find(&friendships)

	// 2. Formatting Output dengan Equipped Items
	var friendList []FriendResponse // Gunakan struct response custom

	for _, f := range friendships {
		var friendData models.User
		
		// Tentukan mana yang 'Teman' (bukan diri sendiri)
		if f.UserID == uint(userID) {
			friendData = f.Friend
		} else {
			friendData = f.User
		}

		// === LOGIC BARU: Ambil Item yang dipakai teman ===
		var equippedItems []models.Item
		config.DB.Table("items").
			Joins("JOIN user_items ON user_items.item_id = items.id").
			Where("user_items.user_id = ? AND user_items.is_equipped = ?", friendData.ID, true).
			Find(&equippedItems)

		// Append ke list response
		friendList = append(friendList, FriendResponse{
			ID:            friendData.ID,
			Name:          friendData.Name,
			Username:      friendData.Username,
			Level:         friendData.Level,
			EquippedItems: equippedItems, // Masukkan data frame/title
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Friends retrieved", friendList)
}


func GetFriendRequests(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	var requests []models.Friendship


	if err := config.DB.Preload("User").
		Where("friend_id = ? AND status = 'pending'", myID).
		Find(&requests).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch requests", nil)
	}

	
	type UserWithItems struct {
		models.User
		EquippedItems []models.Item `json:"equipped_items"`
	}

	type RequestResponse struct {
		ID        uint          `json:"id"` 
		Status    string        `json:"status"`
		CreatedAt time.Time     `json:"created_at"`
		User      UserWithItems `json:"user"` 
	}

	var responseList []RequestResponse

	
	for _, req := range requests {
		var items []models.Item

		
		config.DB.Table("items").
			Joins("JOIN user_items ON user_items.item_id = items.id").
			Where("user_items.user_id = ? AND user_items.is_equipped = ?", req.UserID, true).
			Find(&items)

		responseList = append(responseList, RequestResponse{
			ID:        req.ID,
			Status:    req.Status,
			CreatedAt: req.CreatedAt,
			User: UserWithItems{
				User:          req.User,
				EquippedItems: items,
			},
		})
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Pending requests", responseList)
}


func RemoveFriend(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	targetID := c.Params("id") 

	// Hapus hubungan dua arah
	result := config.DB.Where(
		"((user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)) AND status = 'accepted'",
		myID, targetID, targetID, myID,
	).Delete(&models.Friendship{})

	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Friend not found", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Friend removed", nil)
}

// 7. Get Sent Requests (Lihat siapa saja yang saya add tapi belum di-acc)
func GetSentRequests(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	var requests []models.Friendship

	// Cari yang UserID-nya SAYA, status pending, dan preload data 'Friend' (orang yang dituju)
	config.DB.Preload("Friend").
		Where("user_id = ? AND status = 'pending'", myID).
		Find(&requests)

	return utils.SuccessResponse(c, fiber.StatusOK, "Sent requests retrieved", requests)
}

// 8. Cancel Request (Batalkan permintaan yang sudah dikirim)
func CancelRequest(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	targetID := c.Params("id") // ID user yang mau dibatalkan request-nya

	// Hapus hanya jika status masih 'pending' dan pengirimnya adalah saya
	result := config.DB.Where(
		"user_id = ? AND friend_id = ? AND status = 'pending'",
		myID, targetID,
	).Delete(&models.Friendship{})

	if result.RowsAffected == 0 {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Request not found or already accepted", nil)
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Request cancelled", nil)
}
