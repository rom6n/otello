package httputils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message,omitempty" example:"failed to find hotels"`
	Error   string `json:"error,omitempty" example:"failed to parse query value 'stars-from'"`
}

type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message,omitempty" example:"successfully found hotels"`
	Data    interface{} `json:"data,omitempty"`
}

func HandleUnsuccess(c *fiber.Ctx, message, err string, data interface{}, status int) error {
	return c.Status(status).JSON(&Response{
		Success: false,
		Message: message,
		Error:   err,
		Data:    data,
	})
}

func HandleSuccess(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(&Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func ParseTimeZ(t string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", t)
}

func ParseTimeDate(t string) (time.Time, error) {
	return time.Parse("2006-01-02", t)
}

func QueryOneOf(c *fiber.Ctx, keys ...string) string {
	for _, k := range keys {
		if v := c.Query(k); v != "" {
			return v
		}
	}
	return ""
}
