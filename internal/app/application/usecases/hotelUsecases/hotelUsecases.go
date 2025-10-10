package hotelusecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	hotelrepository "github.com/rom6n/otello/internal/app/adapters/repository/hotelrepository"
	"github.com/rom6n/otello/internal/app/domain/hotel"
)

type HotelUsecases interface {
	Create(ctx context.Context, hotel *hotel.Hotel) error
	Update(ctx context.Context, newHotelData *hotel.Hotel) error
	Delete(ctx context.Context, hotelUuid uuid.UUID) error
	Get(ctx context.Context, hotelUuid uuid.UUID) (*hotel.Hotel, error)
}

type hotelUsecase struct {
	hotelRepo hotelrepository.HotelRepository
	timeout   time.Duration
}

func New(hotelRepo hotelrepository.HotelRepository, timeout time.Duration) HotelUsecases {
	return &hotelUsecase{
		hotelRepo: hotelRepo,
		timeout:   timeout,
	}
}

func (v *hotelUsecase) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *hotelUsecase) Create(ctx context.Context, hotel *hotel.Hotel) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.hotelRepo.CreateHotel(usecaseCtx, hotel)
	if err != nil {
		return err
	}

	return nil
}

func (v *hotelUsecase) Update(ctx context.Context, newHotelData *hotel.Hotel) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.hotelRepo.UpdateHotel(usecaseCtx, newHotelData)
	if err != nil {
		return err
	}

	return nil
}

func (v *hotelUsecase) Delete(ctx context.Context, hotelUuid uuid.UUID) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.hotelRepo.DeleteHotel(usecaseCtx, hotelUuid)
	if err != nil {
		return err
	}

	return nil
}

func (v *hotelUsecase) Get(ctx context.Context, hotelUuid uuid.UUID) (*hotel.Hotel, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	hotel, err := v.hotelRepo.GetHotel(usecaseCtx, hotelUuid)
	if err != nil {
		return nil, err
	}

	return hotel, nil
}
