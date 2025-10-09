package http

import (
	"github.com/gofiber/fiber/v2"
	"github.com/rom6n/otello/internal/app/adapters/handler"
	"github.com/rom6n/otello/internal/app/config"
)

func NewFiberApp(cfg config.Config) *fiber.App {
	app := fiber.New()

	api := app.Group("/api")
	userApi := api.Group("/user")

	userHandler := handler.UserHandler{
		RegisterUsecase: cfg.RegisterUserRepo,
	}

	userApi.Get("/register", userHandler.Register()) // POST сделать потом !!!!!!!!!!!!!
	//userApi.Post("/login", userHandler.Login())           // all data in queries
	//userApi.Put("/change-name", userHandler.ChangeName()) // all data in queries

	return app
}
