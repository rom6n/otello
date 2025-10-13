package config

import (
	"time"

	"github.com/rom6n/otello/internal/app/adapters/repository/flightticketrepository"
	"github.com/rom6n/otello/internal/app/adapters/repository/hotelrepository"
	hotelroomrepository "github.com/rom6n/otello/internal/app/adapters/repository/hotelroomrepository"
	"github.com/rom6n/otello/internal/app/adapters/repository/rentrepository"
	"github.com/rom6n/otello/internal/app/adapters/repository/userrepository"
	"github.com/rom6n/otello/internal/app/application/usecases/flightticketusecases"
	"github.com/rom6n/otello/internal/app/application/usecases/hotelroomusecases"
	"github.com/rom6n/otello/internal/app/application/usecases/hotelusecases"
	"github.com/rom6n/otello/internal/app/application/usecases/rentusecases"
	"github.com/rom6n/otello/internal/app/application/usecases/userusecases"
	"github.com/rom6n/otello/internal/utils/jwtutils"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Config struct {
	UserRepo             userrepository.UserRepository
	JWTRepo              jwtutils.JwtRepository
	UserUsecases         userusecases.UserUsecases
	HotelUsecases        hotelusecases.HotelUsecases
	HotelRoomUsecases    hotelroomusecases.HotelRoomUsecases
	RentUsecases         rentusecases.RentUsecases
	FlightTicketUsecases flightticketusecases.FlightTicketUsecases
}

const DBName = "otello"

func GetConfig(dbClient *mongo.Client) Config {
	userRepo := userrepository.New(dbClient, DBName, "users", 30*time.Second)
	jwtRepo := jwtutils.New()
	hotelRepo := hotelrepository.New(dbClient, DBName, "hotels", 30*time.Second)
	hotelRoomRepo := hotelroomrepository.New(dbClient, DBName, "hotelRooms", 30*time.Second)
	rentRepo := rentrepository.New(dbClient, DBName, "rents", 30*time.Second)
	flightTicketRepo := flightticketrepository.New(dbClient, DBName, "flightTickets", 30*time.Second)

	userUsecase := userusecases.New(userRepo, jwtRepo, 30*time.Second)
	hotelUsecase := hotelusecases.New(hotelRepo, 30*time.Second)
	hotelRoomUsecase := hotelroomusecases.New(hotelRoomRepo, rentRepo, 30*time.Second)
	rentUsecase := rentusecases.New(rentRepo, 30*time.Second)
	flightTicketUsecase := flightticketusecases.New(flightTicketRepo, 30*time.Second)

	return Config{
		UserRepo:             userRepo,
		JWTRepo:              jwtRepo,
		UserUsecases:         userUsecase,
		HotelUsecases:        hotelUsecase,
		HotelRoomUsecases:    hotelRoomUsecase,
		RentUsecases:         rentUsecase,
		FlightTicketUsecases: flightTicketUsecase,
	}
}
