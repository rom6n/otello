package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rom6n/otello/internal/app/adapters/handler"
	"github.com/rom6n/otello/internal/app/config"
	"github.com/rom6n/otello/internal/app/domain/user"
	jwtutils "github.com/rom6n/otello/internal/utils/jwtutils"
)

func NewFiberApp(cfg config.Config) *fiber.App {
	app := fiber.New()

	app.Use(logger.New())

	api := app.Group("/api")
	userApi := api.Group("/user")
	hotelApi := api.Group("/hotel")
	hotelRoomApi := api.Group("/hotel-room")

	adminApi := api.Group("/admin")
	adminHotelApi := adminApi.Group("/hotel")
	adminHotelRoomApi := adminApi.Group("/hotel-room")

	userHandler := handler.UserHandler{
		UserUsecase: cfg.UserUsecases,
	}

	hotelHandler := handler.HotelHandler{
		HotelUsecase: cfg.HotelUsecases,
	}

	hotelRoomHandler := handler.HotelRoomHandler{
		HotelRoomUsecase: cfg.HotelRoomUsecases,
	}

	userApi.Get("/register", userHandler.Register())                                         // POST сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴🔴🔴
	userApi.Get("/login", userHandler.Login())                                               // POST сделать потом !!!!!!!!!!!!!  🔴🔴🔴
	userApi.Get("/rename", CheckJwtMiddleware(cfg.JWTREpo, false), userHandler.ChangeName()) // PUT сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴

	hotelApi.Get("/find", hotelHandler.Find())

	hotelRoomApi.Get("/find", hotelRoomHandler.Find())

	// 5da2255a-1ce7-4427-ad44-862165ebf9d7
	adminHotelApi.Get("/create", CheckJwtMiddleware(cfg.JWTREpo, true), hotelHandler.Create()) // POST сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	adminHotelApi.Get("/update", CheckJwtMiddleware(cfg.JWTREpo, true), hotelHandler.Update()) // PUT сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	adminHotelApi.Get("/delete", CheckJwtMiddleware(cfg.JWTREpo, true), hotelHandler.Delete()) // DELETE сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴

	adminHotelRoomApi.Get("/create", CheckJwtMiddleware(cfg.JWTREpo, true), hotelRoomHandler.Create()) // POST сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	adminHotelRoomApi.Get("/update", CheckJwtMiddleware(cfg.JWTREpo, true), hotelRoomHandler.Update()) // PUT сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	adminHotelRoomApi.Get("/delete", CheckJwtMiddleware(cfg.JWTREpo, true), hotelRoomHandler.Delete()) // DELETE сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴

	return app
}

func CheckJwtMiddleware(jwtRepo jwtutils.JwtRepository, adminOnly bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jwtToken := c.Cookies("jwtToken")
		if jwtToken == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(handler.Response{
				Success: false,
				Message: "login before be proceeed",
				Error:   "unauthorized",
			})
		}

		claims, err := jwtRepo.VerifyJwt(jwtToken)
		if err != nil || claims["iss"] != "otello" {
			return c.Status(fiber.StatusUnauthorized).JSON(handler.Response{
				Success: false,
				Message: "jwt token not verified or accepted",
				Error:   fmt.Sprintf("%v-%v", err, claims["iss"]),
			})
		}

		if adminOnly && claims["role"] != user.RoleAdmin {
			return c.Status(fiber.StatusUnauthorized).JSON(handler.Response{
				Success: false,
				Message: "no permission",
			})
		}

		c.Locals("id", claims["user_id"])
		c.Locals("role", claims["role"])
		flog.Info("MIDDLEWARE PASSED IP: ", c.IP())

		return c.Next()
	}
}
