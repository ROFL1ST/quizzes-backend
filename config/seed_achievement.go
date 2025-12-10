package config

import (
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/models"
)


func SeedAchievements() {
	var count int64
	DB.Model(&models.Achievement{}).Count(&count)

	if count > 0 {
		return
	}

	fmt.Println("Seeding Achievements...")

	achievements := []models.Achievement{
		{ID: 1, Name: "Langkah Pertama", Description: "Menyelesaikan kuis pertama kali", IconURL: "🎯"},
		{ID: 2, Name: "Sempurna!", Description: "Mendapatkan nilai 100 dalam kuis", IconURL: "💯"},
		{ID: 3, Name: "Raja Kuis", Description: "Menyelesaikan 10 kuis", IconURL: "👑"},
		{ID: 4, Name: "Petarung", Description: "Memenangkan duel pertama", IconURL: "⚔️"},
		{ID: 5, Name: "Sepuh", Description: "Mencapai Level 5", IconURL: "👴"},
        {ID: 6, Name: "Konsisten", Description: "Streak selama 3 hari", IconURL: "🔥"},
	}

	for _, a := range achievements {
		DB.Create(&a)
	}
	fmt.Println("Achievements Seeded!")
}