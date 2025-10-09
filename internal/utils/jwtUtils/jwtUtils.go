package jwtutils

import (
	"fmt"
	"log"
	"os"
	"time"

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
		"user_id": uuid,
		"exp":     time.Now().Add(time.Minute * 10).Unix(),
		"iat":     time.Now().Unix(),
		"iss":     "otello",
		"role":    role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(v.signKey)
}

func (v *JwtRepo) VerifyJwt(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return v.signKey, nil
	})

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid || claims["iss"] != "otello" {
		return nil, fmt.Errorf("jwt is not valid: %v-%v-%v", ok, token.Valid, claims["iss"])
	}

	return claims, err
}
