package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/ROFL1ST/quizzes-backend/config"
	"github.com/ROFL1ST/quizzes-backend/utils"
	"github.com/gofiber/fiber/v2"
)

type ServiceStatus struct {
	Name    string `json:"name"`
	Status  string `json:"status"` // "online", "offline", "degraded"
	Latency string `json:"latency"`
	Message string `json:"message,omitempty"`
}

func GetSystemHealth(c *fiber.Ctx) error {
	services := []ServiceStatus{}

	// 1. Check Database
	start := time.Now()
	dbStatus := ServiceStatus{Name: "Database (PostgreSQL)", Status: "online"}
	sqlDB, err := config.DB.DB()
	if err != nil {
		dbStatus.Status = "offline"
		dbStatus.Message = "Connection Error"
	} else {
		if err := sqlDB.Ping(); err != nil {
			dbStatus.Status = "offline"
			dbStatus.Message = "Ping Failed"
		}
	}
	dbStatus.Latency = time.Since(start).String()
	services = append(services, dbStatus)

	// 2. Check ML Service
	startML := time.Now()
	mlStatus := ServiceStatus{Name: "AI Service (ML)", Status: "offline"}
	mlURL := os.Getenv("ML_SERVICE_URL")
	if mlURL == "" {
		mlURL = "http://localhost:5002"
	}

	// Simple HTTP GET to ML Service Root
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(mlURL + "/")
	if err == nil && resp.StatusCode == 200 {
		mlStatus.Status = "online"
	} else {
		mlStatus.Message = "Unreachable"
	}
	mlStatus.Latency = time.Since(startML).String()
	services = append(services, mlStatus)

	// 3. Backend Self (Always Online if this runs)
	services = append(services, ServiceStatus{
		Name:    "Backend API (Go)",
		Status:  "online",
		Latency: "0ms",
	})

	return utils.SuccessResponse(c, fiber.StatusOK, "System Health Status", services)
}
