package httputils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type CookieUsage int

const (
	JwtAccessToken  CookieUsage = 1
	JwtRefreshToken CookieUsage = 2
)

const SecureNeed = false

func BuildCookie(value string, usage CookieUsage) *fiber.Cookie {
	var name string
	var maxAge int
	var expires time.Time
	var httpOnly bool

	switch usage {
	case JwtAccessToken:
		maxAge = 600
		expires = time.Now().Add(10 * time.Minute)
		httpOnly = true
		name = "jwtToken"
	case JwtRefreshToken:
		maxAge = 7 * 24 * 3600
		expires = time.Now().Add(7 * 24 * 3600 * time.Second)
		httpOnly = true
		name = "jwtRefreshToken"
	}

	return &fiber.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Expires:  expires,
		Secure:   SecureNeed,
		SameSite: fiber.CookieSameSiteStrictMode,
		HTTPOnly: httpOnly,
	}
}
