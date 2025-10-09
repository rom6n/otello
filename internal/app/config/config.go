package config

import (
	"time"

	userrepository "github.com/rom6n/otello/internal/app/adapters/repository/userRepository"
	registeruser "github.com/rom6n/otello/internal/app/application/usecases/registerUser"
	jwtutils "github.com/rom6n/otello/internal/utils/jwtUtils"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Config struct {
	UserRepo         userrepository.UserRepository
	JWTREpo          jwtutils.JwtRepository
	RegisterUserRepo registeruser.RegisterUserRepository
}

const DBName = "otello"

func GetConfig(dbClient *mongo.Client) Config {
	userRepo := userrepository.New(dbClient, DBName, "users", 30*time.Second)
	jwtRepo := jwtutils.New()
	registerUserREpo := registeruser.New(userRepo, jwtRepo, 30*time.Second)
	return Config{
		UserRepo:         userRepo,
		JWTREpo:          jwtRepo,
		RegisterUserRepo: registerUserREpo,
	}
}
