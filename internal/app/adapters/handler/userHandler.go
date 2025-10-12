package handler

import (
	"fmt"
	"regexp"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	userusecases "github.com/rom6n/otello/internal/app/application/usecases/userusecases"
	"github.com/rom6n/otello/internal/app/domain/user"
)

type UserHandler struct {
	UserUsecase userusecases.UserUsecases
}

func IsEmailCorrect(email string) bool {
	regex := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(regex)
	return re.MatchString(email)
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
				Error:   "query values 'name', 'email' and 'password' are required",
			})
		}

		if !IsEmailCorrect(email) {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to register",
				Error:   "invalid email",
			})
		}

		newUser := user.NewUser(name, email, password)

		jwtCookie, newUser, err := v.UserUsecase.Register(ctx, newUser)
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
			Data:    newUser,
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

		if !IsEmailCorrect(email) {
			return c.Status(fiber.StatusBadRequest).JSON(Response{
				Success: false,
				Message: "failed to login",
				Error:   "invalid email",
			})
		}

		jwtCookie, foundUser, err := v.UserUsecase.Login(ctx, email, password)
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
			Data:    foundUser,
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
			flog.Warnf("Error parsing user UUID from JWT: %v\n", parseErr)
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
