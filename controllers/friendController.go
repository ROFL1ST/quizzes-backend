package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	
	"github.com/gofiber/fiber/v2"
)

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
	utils.SendNotification(friend.ID, "👋 Permintaan teman baru dari "+input.Username, "/friends", "info")
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
	// Cari request dimana SAYA adalah FriendID (penerima) dan statusnya pending
	if err := config.DB.Where("user_id = ? AND friend_id = ? AND status = 'pending'", input.RequesterID, myID).First(&friendship).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Friend request not found", nil)
	}

	// Update status jadi accepted
	friendship.Status = "accepted"
	config.DB.Save(&friendship)

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
	
	// Ambil data dimana user terlibat (baik sebagai pengirim atau penerima) DAN status accepted
	config.DB.Preload("Friend").Preload("User").
		Where("(user_id = ? OR friend_id = ?) AND status = 'accepted'", userID, userID).
		Find(&friendships)

	// Formatting output agar rapi (hanya menampilkan list teman)
	var friendList []models.User
	for _, f := range friendships {
		if f.UserID == uint(userID) {
			friendList = append(friendList, f.Friend) // Temannya adalah 'Friend'
		} else {
			friendList = append(friendList, f.User)   // Temannya adalah 'User' (requester)
		}
	}

	return utils.SuccessResponse(c, fiber.StatusOK, "Friends retrieved", friendList)
}

// 5. Get Pending Requests (Lihat siapa yang minta berteman ke saya)
func GetFriendRequests(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	var requests []models.Friendship

	// Cari yang FriendID-nya SAYA dan status pending
	config.DB.Preload("User"). // Load data si pengirim (Requester)
		Where("friend_id = ? AND status = 'pending'", myID).
		Find(&requests)

	return utils.SuccessResponse(c, fiber.StatusOK, "Pending requests", requests)
}

// 6. Delete Friend (Menghapus teman yang sudah accepted)
func RemoveFriend(c *fiber.Ctx) error {
	myID := c.Locals("user_id").(float64)
	targetID := c.Params("id") // ID teman yang mau dihapus

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