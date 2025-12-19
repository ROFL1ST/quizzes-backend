// controllers/shopController.go
package controllers

import (
	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/models"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)


func GetShopItems(c *fiber.Ctx) error {
	var items []models.Item

	if err := config.DB.Where("is_active = ?", true).Find(&items).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to fetch items", nil)
	}
	return utils.SuccessResponse(c, fiber.StatusOK, "Shop items retrieved", items)
}

// BuyItem: Membeli barang
func BuyItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64) // Dari JWT Middleware
	var input struct {
		ItemID uint `json:"item_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	// 1. Cek User & Saldo
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "User not found", nil)
	}

	// 2. Cek Barang
	var item models.Item
	if err := config.DB.First(&item, input.ItemID).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusNotFound, "Item not found", nil)
	}

	// 3. Cek apakah sudah punya
	var existingItem models.UserItem
	if err := config.DB.Where("user_id = ? AND item_id = ?", userID, input.ItemID).First(&existingItem).Error; err == nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "You already own this item", nil)
	}

	// 4. Cek Cukup Uang
	if user.Coins < item.Price {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Not enough coins", nil)
	}

	// 5. Transaksi (Kurangi Koin & Tambah Item)
	tx := config.DB.Begin()

	user.Coins -= item.Price
	if err := tx.Save(&user).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Transaction failed", nil)
	}

	userItem := models.UserItem{
		UserID: user.ID,
		ItemID: item.ID,
	}
	if err := tx.Create(&userItem).Error; err != nil {
		tx.Rollback()
		return utils.ErrorResponse(c, fiber.StatusInternalServerError, "Failed to add item", nil)
	}

	tx.Commit()
	utils.CheckDailyMissions(uint(userID), "shop", 1, "buy")
	return utils.SuccessResponse(c, fiber.StatusOK, "Item purchased successfully", fiber.Map{
		"coins_left": user.Coins,
		"item":       item,
	})
}

// GetMyInventory: Melihat barang yang sudah dibeli
func GetMyInventory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var myItems []models.UserItem
	
	config.DB.Preload("Item").Where("user_id = ?", userID).Find(&myItems)
	
	return utils.SuccessResponse(c, fiber.StatusOK, "Inventory retrieved", myItems)
}

func EquipItem(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(float64)
	var input struct {
		ItemID uint `json:"item_id"`
	}

	if err := c.BodyParser(&input); err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid input", nil)
	}

	// 1. Cek apakah user MEMILIKI item tersebut
	var userItem models.UserItem
	if err := config.DB.Preload("Item").Where("user_id = ? AND item_id = ?", userID, input.ItemID).First(&userItem).Error; err != nil {
		return utils.ErrorResponse(c, fiber.StatusBadRequest, "You don't own this item", nil)
	}

	// 2. Un-equip semua item lain dengan TIPE yang sama (misal: copot frame lama)
	// Query: Update user_items set is_equipped = false where user_id = X and item_id IN (select id from items where type = Y)
	config.DB.Model(&models.UserItem{}).
		Where("user_id = ? AND item_id IN (?)", userID, config.DB.Table("items").Select("id").Where("type = ?", userItem.Item.Type)).
		Update("is_equipped", false)

	// 3. Equip item yang baru dipilih
	userItem.IsEquipped = true
	config.DB.Save(&userItem)
	utils.CheckDailyMissions(uint(userID), "shop", 1, "equip")
	return utils.SuccessResponse(c, fiber.StatusOK, "Item equipped successfully", userItem.Item)
}