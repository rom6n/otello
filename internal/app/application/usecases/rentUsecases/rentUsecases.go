package rentusecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	rentrepository "github.com/rom6n/otello/internal/app/adapters/repository/rentrepository"
	"github.com/rom6n/otello/internal/app/domain/rent"
	"github.com/rom6n/otello/internal/app/domain/user"
)

type RentUsecases interface {
	Create(ctx context.Context, rent *rent.Rent) error
	Delete(ctx context.Context, rentUuid, userUuid uuid.UUID, userRole string) error
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

	foundedRents, getRentsErr := v.rentRepo.GetRentsByHotelRoomUuid(usecaseCtx, rent.RoomUuid)
	if getRentsErr != nil {
		return getRentsErr
	}

	for _, foundedRent := range foundedRents {
		if foundedRent.DateFrom < rent.DateTo && foundedRent.DateTo > rent.DateFrom {
			return fmt.Errorf("this hotel room is already rented for this time")
		}
	}

	err := v.rentRepo.CreateRent(usecaseCtx, rent)
	if err != nil {
		return err
	}

	return nil
}

func (v *rentUsecase) Delete(ctx context.Context, rentUuid, userUuid uuid.UUID, userRole string) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	foundedRent, findErr := v.rentRepo.GetRent(usecaseCtx, rentUuid)
	if findErr != nil {
		return findErr
	}

	if userRole != string(user.RoleAdmin) && foundedRent.RenterUuid != userUuid {
		return fmt.Errorf("you dont have permission to delete this rent")
	}

	err := v.rentRepo.DeleteRent(usecaseCtx, rentUuid)
	if err != nil {
		return err
	}

	return nil
}
