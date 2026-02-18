package middleware

import (
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// XApiKeyMiddleware enforces the presence of a valid X-API-KEY header
func XApiKeyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip OPTIONS (Preflight)
		if c.Method() == "OPTIONS" {
			return c.Next()
		}

		apiKey := c.Get("X-API-KEY")
		serverKey := os.Getenv("API_KEY")

		if serverKey == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Server misconfiguration: API_KEY not set",
			})
		}

		if apiKey != serverKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid or missing API Key",
			})
		}

		return c.Next()
	}
}

// RateLimiterConfig creates a standard limiter config
func RateLimiterConfig(max int, expiration time.Duration) limiter.Config {
	return limiter.Config{
		Max:        max,
		Expiration: expiration,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  "error",
				"message": "Too many requests. Please try again later.",
			})
		},
	}
}

// Preconfigured Limiters
var AuthLimiter = limiter.New(RateLimiterConfig(5, 1*time.Minute))
var RegisterLimiter = limiter.New(RateLimiterConfig(10, 1*time.Minute))
var GeneralLimiter = limiter.New(RateLimiterConfig(60, 1*time.Minute))

// StrictOriginMiddleware ensures that state-changing requests come from allowed origins
// This effectively prevents CSRF for APIs that don't use SameSite=Strict cookies
func StrictOriginMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Only check state-changing methods
		method := c.Method()
		if method == "GET" || method == "HEAD" || method == "OPTIONS" {
			return c.Next()
		}

		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		if allowedOrigins == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Server misconfiguration: ALLOWED_ORIGINS not set",
			})
		}

		origin := c.Get("Origin")
		referer := c.Get("Referer")

		// Non-browser tools (e.g. Mobile Apps, Postman) might not send Origin/Referer.
		// They are authenticated via X-API-KEY.
		// CSRF attacks are browser-specific, relying on Cookie/Auth persistence.
		// If Origin is present, it MUST match.

		if origin != "" {
			if !isAllowed(origin, allowedOrigins) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Origin not allowed"})
			}
		} else if referer != "" {
			// Basic Referer check (starts with allowed origin)
			// Referer might contain path, so checks if it STARTS with allowed
			if !isAllowedReferer(referer, allowedOrigins) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Referer not allowed"})
			}
		}

		return c.Next()
	}
}

func isAllowed(origin, allowedList string) bool {
	origins := strings.Split(allowedList, ",")
	for _, o := range origins {
		if strings.TrimSpace(o) == origin {
			return true
		}
	}
	return false
}

func isAllowedReferer(referer, allowedList string) bool {
	origins := strings.Split(allowedList, ",")
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if strings.HasPrefix(referer, o) {
			return true
		}
	}
	return false
}
