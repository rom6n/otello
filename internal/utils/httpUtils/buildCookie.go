package httputils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type CookieUsage int

const (
	jwtToken CookieUsage = 1
)

func BuildCookie(value string, usage CookieUsage) *fiber.Cookie {
	var name string
	var maxAge int
	var expires time.Time
	var secure bool
	var httpOnly bool

	switch usage {
	case jwtToken:
		maxAge = 600
		expires = time.Now().Add(10 * time.Minute)
		secure = false
		httpOnly = false
		name = "jwtToken"
	}

	return &fiber.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Expires:  expires,
		Secure:   secure,
		HTTPOnly: httpOnly,
	}
}
