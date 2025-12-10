package config

import (
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/models"
	"golang.org/x/crypto/bcrypt"
)

func SeedDatabase() {
	var count int64
	DB.Model(&models.Role{}).Count(&count)

	if count > 0 {
		return
	}

	fmt.Println("Seeding database...")

	roles := []models.Role{
		{Name: "supervisor", Description: "Super Admin"},
		{Name: "admin", Description: "Operational Admin"},
		{Name: "pengajar", Description: "Instructor"},
	}

	for _, r := range roles {
		DB.Create(&r)
	}

	var spvRole models.Role
	DB.Where("name = ?", "supervisor").First(&spvRole)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), 10)

	superAdmin := models.Admin{
		Name:     "Super Admin",
		Username: "superadmin",
		Password: string(hashed),
		RoleID:   spvRole.ID,
	}

	DB.Create(&superAdmin)
	fmt.Println("Seeding done! User: superadmin | Pass: 123456")
}

