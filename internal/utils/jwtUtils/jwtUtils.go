package jwtutils

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/domain/user"
	httputils "github.com/rom6n/otello/internal/utils/httputils"
)

type JwtRepository interface {
	NewJwt(uuid uuid.UUID, role user.UserRole, usage httputils.CookieUsage) (string, error)
	VerifyJwt(tokenStr string) (jwt.MapClaims, error)
}

type JwtRepo struct {
	signKey []byte
}

func New() JwtRepository {
	signKey := os.Getenv("JWT_KEY")
	if signKey == "" {
		log.Fatalln("JWT_KEY in env is not set")
	}

	return &JwtRepo{
		signKey: []byte(signKey),
	}
}

func (v *JwtRepo) NewJwt(uuid uuid.UUID, role user.UserRole, usage httputils.CookieUsage) (string, error) {
	var exp int64
	switch usage {
	case httputils.JwtAccessToken:
		exp = time.Now().Add(10 * time.Minute).Unix()
	case httputils.JwtRefreshToken:
		exp = time.Now().Add(7 * 24 * 3600 * time.Second).Unix()
	}

	claims := jwt.MapClaims{
		"user_id": uuid.String(),
		"exp":     exp,
		"iat":     time.Now().Unix(),
		"iss":     "otello",
		"aud":     "otello-users",
		"role":    role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(v.signKey)
}

func (v *JwtRepo) VerifyJwt(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return v.signKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("not authorized")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, fmt.Errorf("jwt token has expired")
		}
	}

	if claims["iss"] != "otello" || claims["aud"] != "otello-users" {
		return nil, fmt.Errorf("jwt data is not valid: %v-%v", claims["iss"], claims["aud"])
	}

	return claims, err
}
