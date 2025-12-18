package config

import (
	"fmt"
	"github.com/ROFL1ST/quizzes-backend/models"
)

func SeedShopItems() {
	var count int64
	DB.Model(&models.Item{}).Count(&count)
	if count > 0 {
		fmt.Println("Shop items already seeded, skipping...")
		return
	}

	fmt.Println("Seeding shop items...")

	items := []models.Item{
		// --- AVATAR FRAMES (Aktif) ---
		{Name: "Wooden Frame", Description: "Bingkai kayu sederhana.", Price: 100, Type: "avatar_frame", AssetURL: "frame_wooden", IsActive: true},
		{Name: "Silver Border", Description: "Bingkai perak mengkilap.", Price: 250, Type: "avatar_frame", AssetURL: "frame_silver", IsActive: true},
		{Name: "Golden Glow", Description: "Aura emas pemenang.", Price: 500, Type: "avatar_frame", AssetURL: "frame_gold", IsActive: true},
		{Name: "Neon Blue", Description: "Cahaya neon biru.", Price: 750, Type: "avatar_frame", AssetURL: "frame_neon_blue", IsActive: true},
		{Name: "Neon Red", Description: "Cahaya neon merah.", Price: 750, Type: "avatar_frame", AssetURL: "frame_neon_red", IsActive: true},
		{Name: "Cyber Glitch", Description: "Efek error digital.", Price: 1500, Type: "avatar_frame", AssetURL: "frame_glitch", IsActive: true},
		{Name: "Royal Crown", Description: "Mahkota raja kuis.", Price: 5000, Type: "avatar_frame", AssetURL: "frame_crown", IsActive: true},
		{Name: "Galaxy Swirl", Description: "Keindahan antariksa.", Price: 5000, Type: "avatar_frame", AssetURL: "frame_galaxy", IsActive: true},

		// --- TITLES (Aktif) ---
		{Name: "Rookie", Description: "Pendatang baru.", Price: 50, Type: "title", AssetURL: "Rookie", IsActive: true},
		{Name: "Quiz Enthusiast", Description: "Penggemar kuis.", Price: 200, Type: "title", AssetURL: "Enthusiast", IsActive: true},
		{Name: "Speed Runner", Description: "Si paling cepat.", Price: 500, Type: "title", AssetURL: "⚡ Speedster", IsActive: true},
		{Name: "Brainiac", Description: "Otak encer.", Price: 1000, Type: "title", AssetURL: "🧠 Brainiac", IsActive: true},
		{Name: "Detective", Description: "Teliti mencari jawaban.", Price: 1500, Type: "title", AssetURL: "🕵️ Detective", IsActive: true},
		{Name: "The Professor", Description: "Sang guru besar.", Price: 2500, Type: "title", AssetURL: "👨‍🏫 Professor", IsActive: true},
		{Name: "Quiz God", Description: "Dewa kuis.", Price: 10000, Type: "title", AssetURL: "🌩️ Godlike", IsActive: true},
		{Name: "Sultan", Description: "Pamer koin.", Price: 9999, Type: "title", AssetURL: "💸 Sultan", IsActive: true},

		// --- THEMES (Inactive / Belum Dirilis) ---
		{
			Name: "Midnight Purple", Description: "Tema gelap ungu.", Price: 500, Type: "theme", AssetURL: "theme_purple",
			IsActive: false, // <--- INACTIVE
		},
		{
			Name: "Ocean Breeze", Description: "Segar biru laut.", Price: 500, Type: "theme", AssetURL: "theme_ocean",
			IsActive: false, // <--- INACTIVE
		},
		{
			Name: "Sunset Orange", Description: "Hangat matahari.", Price: 500, Type: "theme", AssetURL: "theme_sunset",
			IsActive: false, // <--- INACTIVE
		},
		{
			Name: "Matrix Green", Description: "Dunia digital.", Price: 1000, Type: "theme", AssetURL: "theme_matrix",
			IsActive: false, // <--- INACTIVE
		},
	}

	for _, item := range items {
		DB.Create(&item)
	}

	fmt.Println("Success! Shop Items seeded (Themes are hidden).")
}
