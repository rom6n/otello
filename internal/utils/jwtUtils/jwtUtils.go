package jwtutils

import (
	"fmt"
	"log"
	"os"
	"time"

	log2 "github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/domain/user"
)

type JwtRepository interface {
	NewJwt(uuid uuid.UUID, role user.UserRole) (string, error)
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

func (v *JwtRepo) NewJwt(uuid uuid.UUID, role user.UserRole) (string, error) {
	claims := jwt.MapClaims{
		"user_id": uuid.String(),
		"exp":     time.Now().Add(time.Minute * 10).Unix(),
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
		log2.Warnf("failed to parse jwt token: %v", err)
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
