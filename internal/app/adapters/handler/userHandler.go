package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	registeruser "github.com/rom6n/otello/internal/app/application/usecases/registerUser"
	"github.com/rom6n/otello/internal/app/domain/user"
)

type UserHandler struct {
	RegisterUsecase registeruser.RegisterUserRepository // Добавление в БД пользователя
}

func (v *UserHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		name := c.Query("name")
		email := c.Query("email")
		password := c.Query("password")

		if name == "" || email == "" || password == "" {
			return c.JSON(Response{
				Success: false,
				Message: "failed to register",
				Error:   "name, email and password are required",
			})
		}

		user := user.NewUser(name, email, password)

		jwtCookie, user, err := v.RegisterUsecase.Register(ctx, user)
		if err != nil {
			return c.JSON(Response{
				Success: false,
				Message: "failed to register",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		c.Cookie(jwtCookie)

		return c.JSON(Response{
			Success: true,
			Message: "successfully registered",
			Data:    map[string]string{"Name": name, "Email": email, "Hashed password": user.Password, "Role": string(user.Role)},
		})
	}
}

func (v *UserHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}

func (v *UserHandler) ChangeName() fiber.Handler {
	return func(c *fiber.Ctx) error {
		return nil
	}
}
