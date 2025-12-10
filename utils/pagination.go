package utils

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PaginationParams struct {
	Page     int
	PageSize int
	Offset   int
}

type PaginatedResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    MetaData    `json:"meta"`
}

type MetaData struct {
	CurrentPage int   `json:"current_page"`
	PageSize    int   `json:"page_size"`
	TotalItems  int64 `json:"total_items"`
	TotalPages  int   `json:"total_pages"`
}

// GetPaginationParams mengambil parameter page dan page_size dari query string
func GetPaginationParams(c *fiber.Ctx) PaginationParams {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	pageSize, _ := strconv.Atoi(c.Query("limit", "10"))
	log.Println("Pagination - page:", page, "page_size:", pageSize)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	return PaginationParams{
		Page:     page,
		PageSize: pageSize,
		Offset:   offset,
	}
}

// PaginatedSuccessResponse mengirim response dengan metadata pagination
func PaginatedSuccessResponse(c *fiber.Ctx, statusCode int, message string, data interface{}, total int64, params PaginationParams) error {
	totalPages := int((total + int64(params.PageSize) - 1) / int64(params.PageSize))

	return c.Status(statusCode).JSON(PaginatedResponse{
		Status:  "success",
		Message: message,
		Data:    data,
		Meta: MetaData{
			CurrentPage: params.Page,
			PageSize:    params.PageSize,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	})
}