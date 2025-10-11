package config

import (
	"time"

	hotelrepository "github.com/rom6n/otello/internal/app/adapters/repository/hotelrepository"
	hotelroomrepository "github.com/rom6n/otello/internal/app/adapters/repository/hotelroomrepository"
	rentrepository "github.com/rom6n/otello/internal/app/adapters/repository/rentrepository"
	userrepository "github.com/rom6n/otello/internal/app/adapters/repository/userrepository"
	hotelroomusecases "github.com/rom6n/otello/internal/app/application/usecases/hotelroomusecases"
	hotelusecases "github.com/rom6n/otello/internal/app/application/usecases/hotelusecases"
	rentusecases "github.com/rom6n/otello/internal/app/application/usecases/rentusecases"
	userusecases "github.com/rom6n/otello/internal/app/application/usecases/userusecases"
	jwtutils "github.com/rom6n/otello/internal/utils/jwtutils"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Config struct {
	UserRepo          userrepository.UserRepository
	JWTRepo           jwtutils.JwtRepository
	UserUsecases      userusecases.UserUsecases
	HotelUsecases     hotelusecases.HotelUsecases
	HotelRoomUsecases hotelroomusecases.HotelRoomUsecases
	RentUsecases      rentusecases.RentUsecases
}

const DBName = "otello"

func GetConfig(dbClient *mongo.Client) Config {
	userRepo := userrepository.New(dbClient, DBName, "users", 30*time.Second)
	jwtRepo := jwtutils.New()
	hotelRepo := hotelrepository.New(dbClient, DBName, "hotels", 30*time.Second)
	hotelRoomRepo := hotelroomrepository.New(dbClient, DBName, "hotelRooms", 30*time.Second)
	rentRepo := rentrepository.New(dbClient, DBName, "rents", 30*time.Second)

	userUsecase := userusecases.New(userRepo, jwtRepo, 30*time.Second)
	hotelUsecase := hotelusecases.New(hotelRepo, 30*time.Second)
	hotelRoomUsecase := hotelroomusecases.New(hotelRoomRepo, rentRepo, 30*time.Second)
	rentUsecase := rentusecases.New(rentRepo, 30*time.Second)

	return Config{
		UserRepo:          userRepo,
		JWTRepo:           jwtRepo,
		UserUsecases:      userUsecase,
		HotelUsecases:     hotelUsecase,
		HotelRoomUsecases: hotelRoomUsecase,
		RentUsecases:      rentUsecase,
	}
}
