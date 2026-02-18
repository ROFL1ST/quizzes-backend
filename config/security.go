package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2/middleware/cors"
)

const defaultAllowedOrigins = "https://quizapp-indo.vercel.app, https://planetpulse-admin-gwcx.vercel.app, http://localhost:5173, http://localhost:3000"

func AllowedOrigins() string {
	origins := strings.TrimSpace(os.Getenv("CORS_ALLOW_ORIGINS"))
	if origins == "" {
		return defaultAllowedOrigins
	}
	return origins
}

func BuildCORSConfig() cors.Config {
	return cors.Config{
		AllowOrigins:     AllowedOrigins(),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-CSRF-Token",
		AllowCredentials: true,
	}
}

func IsAllowedOrigin(origin string) bool {
	origin = strings.TrimSpace(origin)
	if origin == "" {
		return false
	}

	for _, allowed := range strings.Split(AllowedOrigins(), ",") {
		if strings.TrimSpace(allowed) == origin {
			return true
		}
	}

	return false
}

func ValidateCriticalConfig() error {
	secret := strings.TrimSpace(os.Getenv("JWT_SECRET"))
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET is required and must be at least 32 characters")
	}

	return nil
}
