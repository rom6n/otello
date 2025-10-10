package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	userusecases "github.com/rom6n/otello/internal/app/application/usecases/userusecases"
	"github.com/rom6n/otello/internal/app/domain/user"
)

type UserHandler struct {
	UserUsecase userusecases.UserUsecases
}

func (v *UserHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		name := c.Query("name")
		email := c.Query("email")
		password := c.Query("password")

		if name == "" || email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to register",
				Error:   "name, email and password are required",
			})
		}

		user := user.NewUser(name, email, password)

		jwtCookie, user, err := v.UserUsecase.Register(ctx, user)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
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
		ctx := c.Context()
		email := c.Query("email")
		password := c.Query("password")

		if email == "" || password == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to login",
				Error:   "email and password are required",
			})
		}

		jwtCookie, err := v.UserUsecase.Login(ctx, email, password)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to login",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		c.Cookie(jwtCookie)

		return c.JSON(Response{
			Success: true,
			Message: "successfully logged in",
		})
	}
}

func (v *UserHandler) ChangeName() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		newName := c.Query("new-name")

		userIdStr := c.Locals("id").(string)
		userId, parseErr := uuid.Parse(userIdStr)
		if parseErr != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to change a name",
				Error:   fmt.Sprintf("uuid parse error: %v", parseErr),
			})
		}

		if newName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to change a name",
				Error:   "dont forget to enter a new name",
			})
		}

		if err := v.UserUsecase.ChangeName(ctx, userId, newName); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(Response{
				Success: false,
				Message: "failed to change a name",
				Error:   fmt.Sprintf("%v", err),
			})
		}

		return c.JSON(Response{
			Success: true,
			Message: fmt.Sprintf("successfully changed name to '%v'", newName),
		})
	}
}
