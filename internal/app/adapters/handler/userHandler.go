package handler

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
	flog "github.com/gofiber/fiber/v2/log"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/application/usecases/userusecases"
	"github.com/rom6n/otello/internal/app/domain/user"
	"github.com/rom6n/otello/internal/utils/httputils"
)

type UserHandler struct {
	UserUsecase userusecases.UserUsecases
}

// @Summary Зарегистрироваться
// @Description Регистрирует пользователя по email, имени и паролю
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param name query string true "Имя пользователя"
// @Param email query string true "Email пользователя"
// @Param password query string true "Пароль пользователя"
// @Success 200 {object} httputils.SuccessResponse{data=user.User}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/user/register [post]
func (v *UserHandler) Register() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		name := c.Query("name")
		email := c.Query("email")
		password := c.Query("password")

		unsuccessMessage := "failed to register"

		if name == "" || email == "" || password == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query values 'name', 'email' and 'password' are required", nil, fiber.StatusBadRequest)
		}

		if !httputils.IsEmailCorrect(email) {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "invalid email address", nil, fiber.StatusBadRequest)
		}

		newUser := user.NewUser(name, email, password)

		jwtRefreshCookie, jwtAccessCookie, err := v.UserUsecase.Register(ctx, newUser)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		c.Cookie(jwtRefreshCookie)
		c.Cookie(jwtAccessCookie)

		return c.JSON(httputils.Response{
			Success: true,
			Message: "successfully registered",
			Data:    newUser,
		})
	}
}

// @Summary Войти в аккаунт
// @Description Вход в аккаунт по email и паролю
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param email query string true "Email"
// @Param password query string true "Пароль"
// @Success 200 {object} httputils.SuccessResponse{data=user.User}
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/user/login [post]
func (v *UserHandler) Login() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		email := c.Query("email")
		password := c.Query("password")

		unsuccessMessage := "failed to login"

		if email == "" || password == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query values 'email' and 'password' are required", nil, fiber.StatusBadRequest)
		}

		if !httputils.IsEmailCorrect(email) {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "invalid email address", nil, fiber.StatusBadRequest)
		}

		jwtRefreshCookie, jwtAccessCookie, foundUser, err := v.UserUsecase.Login(ctx, email, password)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		c.Cookie(jwtRefreshCookie)
		c.Cookie(jwtAccessCookie)

		return c.JSON(httputils.Response{
			Success: true,
			Message: "successfully logged in",
			Data:    foundUser,
		})
	}
}

// @Summary Изменить имя
// @Description Изменяет имя пользователя. Требуется авторизация
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param name query string true "Новое имя пользователя"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/user/rename [put]
func (v *UserHandler) ChangeName() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		newName := c.Query("name")

		unsuccessMessage := "failed to change the name"

		userIdStr := c.Locals("id").(string)
		userId, parseUuidErr := uuid.Parse(userIdStr)
		if parseUuidErr != nil {
			flog.Warnf("Error parsing user UUID from Locals: %v\n", parseUuidErr)
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("uuid parse error: %v", parseUuidErr), nil, fiber.StatusInternalServerError)
		}

		if newName == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "new name is required", nil, fiber.StatusBadRequest)
		}

		if err := v.UserUsecase.ChangeName(ctx, userId, newName); err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		return c.JSON(httputils.Response{
			Success: true,
			Message: fmt.Sprintf("successfully changed the name to '%v'", newName),
		})
	}
}

// @Summary Выдать роль админа
// @Description Выдает роль админа по паролю. Требуется авторизация
// @Tags Пользователь
// @Accept json
// @Produce json
// @Param password query string true "Пароль админки"
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/user/get-admin [post]
func (v *UserHandler) GetAdminRole() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		password := c.Query("password")

		unsuccessMessage := "failed to get admin role"

		if password == "" {
			return httputils.HandleUnsuccess(c, unsuccessMessage, "query value 'password' is required", nil, fiber.StatusBadRequest)
		}

		userUuidStr := c.Locals("id").(string)
		userId, parseErr := uuid.Parse(userUuidStr)
		if parseErr != nil {
			flog.Warnf("Error parsing user UUID from Locals: %v\n", parseErr)
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("uuid parse error: %v", parseErr), nil, fiber.StatusInternalServerError)
		}

		newJwtRefreshCookie, newJwtAccessCookie, err := v.UserUsecase.ChangeRole(ctx, userId, user.RoleAdmin, password)
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		c.Cookie(newJwtRefreshCookie)
		c.Cookie(newJwtAccessCookie)

		return httputils.HandleSuccess(c, "successfully changed role to 'admin'", nil)
	}
}

// @Summary Выдать роль пользователя
// @Description Выдает роль пользователя. Требуется авторизация
// @Tags Пользователь
// @Accept json
// @Produce json
// @Success 200 {object} httputils.SuccessResponse
// @Failure 400 {object} httputils.ErrorResponse
// @Failure 500 {object} httputils.ErrorResponse
// @Router /api/user/get-user [post]
func (v *UserHandler) GetUserRole() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := c.Context()
		unsuccessMessage := "failed to get user role"

		userUuidStr := c.Locals("id").(string)
		userId, parseErr := uuid.Parse(userUuidStr)
		if parseErr != nil {
			flog.Warnf("Error parsing user UUID from Locals: %v\n", parseErr)
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("uuid parse error: %v", parseErr), nil, fiber.StatusInternalServerError)
		}

		newJwtRefreshCookie, newJwtAccessCookie, err := v.UserUsecase.ChangeRole(ctx, userId, user.RoleUser, "")
		if err != nil {
			return httputils.HandleUnsuccess(c, unsuccessMessage, fmt.Sprintf("%v", err), nil, fiber.StatusInternalServerError)
		}

		c.Cookie(newJwtRefreshCookie)
		c.Cookie(newJwtAccessCookie)

		return httputils.HandleSuccess(c, "successfully changed role to 'user'", nil)
	}
}
