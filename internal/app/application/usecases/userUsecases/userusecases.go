package userusecases

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/adapters/repository/userrepository"
	"github.com/rom6n/otello/internal/app/domain/user"
	"github.com/rom6n/otello/internal/utils/hashutils"
	"github.com/rom6n/otello/internal/utils/httputils"
	"github.com/rom6n/otello/internal/utils/jwtutils"
)

const AdminPasswordConst = "5SMpOxD4IFrXfCqy.T0jQA4ufiTHnLbLsUJySnL6mgB6u65ZmMjxiTsuXQcI"

type UserUsecases interface {
	Register(ctx context.Context, user *user.User) (*fiber.Cookie, *fiber.Cookie, error)
	Login(ctx context.Context, email, password string) (*fiber.Cookie, *fiber.Cookie, *user.User, error)
	ChangeName(ctx context.Context, userId uuid.UUID, newName string) error
	ChangeRole(ctx context.Context, userId uuid.UUID, newRole user.UserRole, password string) (*fiber.Cookie, *fiber.Cookie, error)
}

type userUsecase struct {
	userRepo     userrepository.UserRepository
	jwtUtilsRepo jwtutils.JwtRepository
	timeout      time.Duration
}

func New(userRepo userrepository.UserRepository, jwtUtilsRepo jwtutils.JwtRepository, timeout time.Duration) UserUsecases {
	return &userUsecase{
		userRepo:     userRepo,
		jwtUtilsRepo: jwtUtilsRepo,
		timeout:      timeout,
	}
}

func (v *userUsecase) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *userUsecase) Register(ctx context.Context, user *user.User) (*fiber.Cookie, *fiber.Cookie, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	salt, saltErr := hashutils.GenerateSalt()
	if saltErr != nil {
		return nil, nil, fmt.Errorf("failed to generate salt: %v", saltErr)
	}

	hashedPassword := hashutils.HashPassword(user.Password, salt)
	user.Password = hashedPassword

	if createErr := v.userRepo.CreateUser(usecaseCtx, user); createErr != nil {
		return nil, nil, createErr
	}

	jwtRefreshToken, jwtErr := v.jwtUtilsRepo.NewJwt(user.Uuid, user.Role, httputils.JwtRefreshToken)
	if jwtErr != nil {
		return nil, nil, fmt.Errorf("failed to create refresh jwt token: %v", jwtErr)
	}

	jwtAccessToken, jwtErr := v.jwtUtilsRepo.NewJwt(user.Uuid, user.Role, httputils.JwtAccessToken)
	if jwtErr != nil {
		return nil, nil, fmt.Errorf("failed to create access jwt token: %v", jwtErr)
	}

	return httputils.BuildCookie(jwtRefreshToken, httputils.JwtRefreshToken), httputils.BuildCookie(jwtAccessToken, httputils.JwtAccessToken), nil
}

func (v *userUsecase) Login(ctx context.Context, email, password string) (*fiber.Cookie, *fiber.Cookie, *user.User, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	foundUser, userErr := v.userRepo.GetUser(usecaseCtx, email)
	if userErr != nil {
		return nil, nil, nil, userErr
	}

	if verifyErr := hashutils.VerifyPassword(password, foundUser.Password); verifyErr != nil {
		return nil, nil, nil, fmt.Errorf("failed to verify password: %v", verifyErr)
	}

	jwtRefreshToken, jwtErr := v.jwtUtilsRepo.NewJwt(foundUser.Uuid, foundUser.Role, httputils.JwtRefreshToken)
	if jwtErr != nil {
		return nil, nil, nil, fmt.Errorf("failed to create refresh jwt token: %v", jwtErr)
	}

	jwtAccessToken, jwtErr := v.jwtUtilsRepo.NewJwt(foundUser.Uuid, foundUser.Role, httputils.JwtAccessToken)
	if jwtErr != nil {
		return nil, nil, nil, fmt.Errorf("failed to create access jwt token: %v", jwtErr)
	}

	return httputils.BuildCookie(jwtRefreshToken, httputils.JwtRefreshToken), httputils.BuildCookie(jwtAccessToken, httputils.JwtAccessToken), foundUser, nil
}

func (v *userUsecase) ChangeName(ctx context.Context, userId uuid.UUID, newName string) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	if err := v.userRepo.UpdateUserName(usecaseCtx, userId, newName); err != nil {
		return err
	}

	return nil
}

func (v *userUsecase) ChangeRole(ctx context.Context, userId uuid.UUID, newRole user.UserRole, password string) (*fiber.Cookie, *fiber.Cookie, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	if newRole == user.RoleAdmin {
		if err := hashutils.VerifyPassword(password, AdminPasswordConst); err != nil {
			return nil, nil, fmt.Errorf("failed to verify password, wrong password: %v", err)
		}
	}

	if err := v.userRepo.UpdateUserRole(usecaseCtx, userId, string(newRole)); err != nil {
		return nil, nil, err
	}

	jwtRefreshToken, jwtErr := v.jwtUtilsRepo.NewJwt(userId, newRole, httputils.JwtRefreshToken)
	if jwtErr != nil {
		return nil, nil, fmt.Errorf("failed to create refresh jwt token: %v", jwtErr)
	}

	return httputils.BuildCookie(jwtRefreshToken, httputils.JwtRefreshToken), httputils.BuildCookie(jwtRefreshToken, httputils.JwtAccessToken), nil
}
