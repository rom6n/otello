package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
	"github.com/rom6n/otello/internal/app/adapters/handler"
	"github.com/rom6n/otello/internal/app/config"
	"github.com/rom6n/otello/internal/app/domain/user"
	httputils "github.com/rom6n/otello/internal/utils/httputils"
	"github.com/rom6n/otello/internal/utils/jwtutils"
)

func NewFiberApp(cfg config.Config) *fiber.App {
	app := fiber.New()
	app.Use(logger.New())

	app.Get("/docs/*", swagger.HandlerDefault)

	api := app.Group("/api")

	userApi := api.Group("/user")
	hotelApi := api.Group("/hotel")
	hotelRoomApi := api.Group("/hotel-room")
	flightTicketApi := api.Group("/flight-ticket")

	adminApi := api.Group("/admin", CheckJwtMiddleware(cfg.JWTRepo, true))
	adminHotelApi := adminApi.Group("/hotel")
	adminHotelRoomApi := adminApi.Group("/hotel-room")
	adminFlightTicketApi := adminApi.Group("/flight-ticket")

	userHandler := handler.UserHandler{
		UserUsecase: cfg.UserUsecases,
	}

	hotelHandler := handler.HotelHandler{
		HotelUsecase: cfg.HotelUsecases,
	}

	hotelRoomHandler := handler.HotelRoomHandler{
		HotelRoomUsecase: cfg.HotelRoomUsecases,
	}

	rentHandler := handler.RentHandler{
		RentUsecase: cfg.RentUsecases,
	}

	flightTicketHandler := handler.FlightTicketHandler{
		FlightTicketUsecase: cfg.FlightTicketUsecases,
	}

	userApi.Post("/register", userHandler.Register())
	userApi.Post("/login", userHandler.Login())
	userApi.Put("/rename", CheckJwtMiddleware(cfg.JWTRepo, false), userHandler.ChangeName())

	hotelApi.Get("/find", hotelHandler.Find())

	hotelRoomApi.Get("/find", hotelRoomHandler.Find())
	hotelRoomApi.Post("/rent", CheckJwtMiddleware(cfg.JWTRepo, false), rentHandler.Create())
	hotelRoomApi.Post("/unrent", CheckJwtMiddleware(cfg.JWTRepo, false), rentHandler.Delete())

	flightTicketApi.Get("/find", flightTicketHandler.Find())
	flightTicketApi.Post("/buy", flightTicketHandler.Buy())

	// 5da2255a-1ce7-4427-ad44-862165ebf9d7
	adminHotelApi.Post("/create", hotelHandler.Create())
	adminHotelApi.Put("/update", hotelHandler.Update())
	adminHotelApi.Delete("/delete", hotelHandler.Delete())

	adminHotelRoomApi.Post("/create", hotelRoomHandler.Create())
	adminHotelRoomApi.Put("/update", hotelRoomHandler.Update())
	adminHotelRoomApi.Delete("/delete", hotelRoomHandler.Delete())

	adminFlightTicketApi.Post("/create", flightTicketHandler.Create())
	adminFlightTicketApi.Put("/update", flightTicketHandler.Update())
	adminFlightTicketApi.Delete("/delete", flightTicketHandler.Delete())

	return app
}

func CheckJwtMiddleware(jwtRepo jwtutils.JwtRepository, adminOnly bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtToken := c.Cookies("jwtToken")
		if jwtToken == "" {
			return httputils.HandleUnsuccess(c, "login before be processed", "unauthorized", nil, fiber.StatusUnauthorized)
		}

		claims, err := jwtRepo.VerifyJwt(jwtToken)
		if err != nil || claims["iss"] != "otello" {
			return httputils.HandleUnsuccess(c, "jwt token not verified or accepted", fmt.Sprintf("%v-%v", err, claims["iss"]), nil, fiber.StatusForbidden)
		}

		fmt.Println(claims["role"])
		if adminOnly && claims["role"] != string(user.RoleAdmin) {
			return httputils.HandleUnsuccess(c, "no permission", "", nil, fiber.StatusForbidden)
		}

		c.Locals("id", claims["user_id"])
		c.Locals("role", claims["role"])
		flog.Info("MIDDLEWARE PASSED IP: ", c.IP())

		return c.Next()
	}
}
