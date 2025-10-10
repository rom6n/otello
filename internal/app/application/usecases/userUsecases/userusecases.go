package userusecases

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	userrepository "github.com/rom6n/otello/internal/app/adapters/repository/userRepository"
	"github.com/rom6n/otello/internal/app/domain/user"
	hashutils "github.com/rom6n/otello/internal/utils/hashUtils"
	httputils "github.com/rom6n/otello/internal/utils/httpUtils"
	jwtutils "github.com/rom6n/otello/internal/utils/jwtUtils"
)

type UserUsecases interface {
	Register(ctx context.Context, user *user.User) (*fiber.Cookie, *user.User, error)
	Login(ctx context.Context, email, password string) (*fiber.Cookie, error)
	ChangeName(ctx context.Context, userId uuid.UUID, newName string) error
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

func (v *userUsecase) Register(ctx context.Context, user *user.User) (*fiber.Cookie, *user.User, error) {
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

	jwtToken, jwtErr := v.jwtUtilsRepo.NewJwt(user.Uuid, user.Role)
	if jwtErr != nil {
		return nil, nil, fmt.Errorf("failed to create jwt token: %v", jwtErr)
	}

	return httputils.BuildCookie(jwtToken, 1), user, nil
}

func (v *userUsecase) Login(ctx context.Context, email, password string) (*fiber.Cookie, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	user, userErr := v.userRepo.GetUser(usecaseCtx, email)
	if userErr != nil {
		return nil, userErr
	}

	if verifyErr := hashutils.VerifyPassword(password, user.Password); verifyErr != nil {
		return nil, fmt.Errorf("failed to verify password: %v", verifyErr)
	}

	jwtToken, jwtErr := v.jwtUtilsRepo.NewJwt(user.Uuid, user.Role)
	if jwtErr != nil {
		return nil, fmt.Errorf("failed to create jwt token: %v", jwtErr)
	}

	return httputils.BuildCookie(jwtToken, 1), nil
}

func (v *userUsecase) ChangeName(ctx context.Context, userId uuid.UUID, newName string) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	if err := v.userRepo.UpdateUserName(usecaseCtx, userId, newName); err != nil {
		return err
	}

	return nil
}
