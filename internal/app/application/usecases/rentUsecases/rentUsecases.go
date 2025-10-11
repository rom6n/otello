package rentusecases

import (
	"context"
	"time"

	rentrepository "github.com/rom6n/otello/internal/app/adapters/repository/rentrepository"
	"github.com/rom6n/otello/internal/app/domain/rent"
)

type RentUsecases interface {
	Create(ctx context.Context, rent *rent.Rent) error
}

type rentUsecase struct {
	rentRepo rentrepository.RentRepository
	timeout  time.Duration
}

func New(rentRepo rentrepository.RentRepository, timeout time.Duration) RentUsecases {
	return &rentUsecase{
		rentRepo: rentRepo,
		timeout:  timeout,
	}
}

func (v *rentUsecase) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *rentUsecase) Create(ctx context.Context, rent *rent.Rent) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.rentRepo.CreateRent(usecaseCtx, rent)
	if err != nil {
		return err
	}

	return nil
}
