package registeruser

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	userrepository "github.com/rom6n/otello/internal/app/adapters/repository/userRepository"
	"github.com/rom6n/otello/internal/app/domain/user"
	hashutils "github.com/rom6n/otello/internal/utils/hashUtils"
	httputils "github.com/rom6n/otello/internal/utils/httpUtils"
	jwtutils "github.com/rom6n/otello/internal/utils/jwtUtils"
)

type RegisterUserRepository interface {
	Register(ctx context.Context, user *user.User) (*fiber.Cookie, *user.User, error)
}

type registerUserRepo struct {
	userRepo     userrepository.UserRepository
	jwtUtilsRepo jwtutils.JwtRepository
	timeout      time.Duration
}

func New(userRepo userrepository.UserRepository, jwtUtilsRepo jwtutils.JwtRepository, timeout time.Duration) RegisterUserRepository {
	return &registerUserRepo{
		userRepo:     userRepo,
		jwtUtilsRepo: jwtUtilsRepo,
		timeout:      timeout,
	}
}

func (v *registerUserRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *registerUserRepo) Register(ctx context.Context, user *user.User) (*fiber.Cookie, *user.User, error) {
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
