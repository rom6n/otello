package http

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rom6n/otello/internal/app/adapters/handler"
	"github.com/rom6n/otello/internal/app/config"
	"github.com/rom6n/otello/internal/app/domain/user"
	jwtutils "github.com/rom6n/otello/internal/utils/jwtUtils"
)

func NewFiberApp(cfg config.Config) *fiber.App {
	app := fiber.New()

	app.Use(logger.New())

	api := app.Group("/api")
	userApi := api.Group("/user")
	//hotelApi := api.Group("/hotel")
	adminApi := api.Group("/admin")
	adminHotelApi := adminApi.Group("/hotel")
	//adminRoomApi := adminApi.Group("/room")

	userHandler := handler.UserHandler{
		UserUsecase: cfg.UserUsecases,
	}

	hotelHandler := handler.HotelHandler{
		HotelUsecase: cfg.HotelUsecases,
	}

	userApi.Get("/register", userHandler.Register())                                         // POST сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴🔴🔴
	userApi.Get("/login", userHandler.Login())                                               // POST сделать потом !!!!!!!!!!!!!  🔴🔴🔴
	userApi.Get("/rename", CheckJwtMiddleware(cfg.JWTREpo, false), userHandler.ChangeName()) // PUT сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴

	adminHotelApi.Get("/create", CheckJwtMiddleware(cfg.JWTREpo, true), hotelHandler.Create()) // POST сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	adminHotelApi.Get("/update", CheckJwtMiddleware(cfg.JWTREpo, true), hotelHandler.Update()) // PUT сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	adminHotelApi.Get("/delete", CheckJwtMiddleware(cfg.JWTREpo, true), hotelHandler.Delete()) // DELETE сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴

	//adminRoomApi.Get("create") // POST сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	//adminRoomApi.Get("update") // PUT сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴
	//adminRoomApi.Get("delete") // DELETE сделать потом !!!!!!!!!!!!! 🔴🔴🔴🔴🔴

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
