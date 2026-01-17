package middleware

import (
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

func Protected() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Method() == "OPTIONS" {
			return c.Next()
		}

		var tokenString string

		authHeader := c.Get("Authorization")
		if authHeader != "" {
			tokenString = strings.Replace(authHeader, "Bearer ", "", 1)
		}

		if tokenString == "" {
			tokenString = c.Cookies("jwt")
		}

		// if tokenString == "" {
		// 	tokenString = c.Query("token")
		// }

		if tokenString == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid Token"})
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid Token"})
		}

		userID, ok := claims["user_id"]
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid Token"})
		}

		role, ok := claims["role"].(string)
		if !ok {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid Token"})
		}

		c.Locals("user_id", userID)
		c.Locals("role", role)

		return c.Next()
	}
}

func AllowRoles(allowedRoles ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		for _, role := range allowedRoles {
			if role == userRole {
				return c.Next()
			}
		}

		return c.Status(403).JSON(fiber.Map{"error": "Forbidden"})
	}
}
