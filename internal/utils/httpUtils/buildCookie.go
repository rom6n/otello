package httputils

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type CookieUsage int

const (
	JwtAccessToken  CookieUsage = 1
	JwtRefreshToken CookieUsage = 2

	jwtAccessTokenMaxAgeSeconds  = 600
	jwtRefreshTokenMaxAgeSeconds = 7 * 24 * 3600

	SecureNeed   = false
	httpOnlyNeed = true
)

func BuildCookie(value string, usage CookieUsage) *fiber.Cookie {
	var name string
	var maxAge int
	var expires time.Time

	switch usage {
	case JwtAccessToken:
		maxAge = jwtAccessTokenMaxAgeSeconds
		expires = time.Now().Add(jwtAccessTokenMaxAgeSeconds * time.Second)
		name = "jwtToken"
	case JwtRefreshToken:
		maxAge = jwtRefreshTokenMaxAgeSeconds
		expires = time.Now().Add(jwtRefreshTokenMaxAgeSeconds * time.Second)
		name = "jwtRefreshToken"
	}

	return &fiber.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Expires:  expires,
		Secure:   SecureNeed,
		SameSite: fiber.CookieSameSiteStrictMode,
		HTTPOnly: httpOnlyNeed,
	}
}
