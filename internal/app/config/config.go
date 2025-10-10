package config

import (
	"time"

	hotelrepository "github.com/rom6n/otello/internal/app/adapters/repository/hotelRepository"
	userrepository "github.com/rom6n/otello/internal/app/adapters/repository/userRepository"
	hotelusecases "github.com/rom6n/otello/internal/app/application/usecases/hotelUsecases"
	userusecases "github.com/rom6n/otello/internal/app/application/usecases/userusecases"
	jwtutils "github.com/rom6n/otello/internal/utils/jwtUtils"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Config struct {
	UserRepo      userrepository.UserRepository
	JWTREpo       jwtutils.JwtRepository
	UserUsecases  userusecases.UserUsecases
	HotelUsecases hotelusecases.HotelUsecases
}

const DBName = "otello"

func GetConfig(dbClient *mongo.Client) Config {
	userRepo := userrepository.New(dbClient, DBName, "users", 30*time.Second)
	jwtRepo := jwtutils.New()
	hotelRepo := hotelrepository.New(dbClient, DBName, "hotels", 30*time.Second)

	userUsecase := userusecases.New(userRepo, jwtRepo, 30*time.Second)
	hotelUsecase := hotelusecases.New(hotelRepo, 30*time.Second)

	return Config{
		UserRepo:      userRepo,
		JWTREpo:       jwtRepo,
		UserUsecases:  userUsecase,
		HotelUsecases: hotelUsecase,
	}
}
