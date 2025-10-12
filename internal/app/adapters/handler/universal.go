package handler

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

func handleUnsuccess(c *fiber.Ctx, message, err string, data interface{}, status int) error {
	return c.Status(status).JSON(&Response{
		Success: false,
		Message: message,
		Error:   err,
		Data:    data,
	})
}

func handleSuccess(c *fiber.Ctx, message string, data interface{}) error {
	return c.Status(fiber.StatusOK).JSON(&Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func parseTimeZ(t string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05Z", t)
}

func queryOneOf(c *fiber.Ctx, keys ...string) string {
	for _, k := range keys {
		if v := c.Query(k); v != "" {
			return v
		}
	}
	return ""
}
