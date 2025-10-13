package hotelusecases

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/adapters/repository/hotelrepository"
	"github.com/rom6n/otello/internal/app/domain/hotel"
)

type HotelUsecases interface {
	Create(ctx context.Context, hotel *hotel.Hotel) error
	Update(ctx context.Context, newHotelData *hotel.Hotel) error
	Delete(ctx context.Context, hotelUuid uuid.UUID) error
	Get(ctx context.Context, hotelUuid uuid.UUID) (*hotel.Hotel, error)
	GetWithParams(ctx context.Context, city string, stars int32, hotelUuid uuid.UUID, needSort, isAsc bool) ([]hotel.Hotel, error)
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

	foundHotel, err := v.hotelRepo.GetHotel(usecaseCtx, hotelUuid)
	if err != nil {
		return nil, err
	}

	return foundHotel, nil
}

func (v *hotelUsecase) GetWithParams(ctx context.Context, city string, stars int32, hotelUuid uuid.UUID, needSort, isAsc bool) ([]hotel.Hotel, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	hotels, err := v.hotelRepo.GetHotelWithParams(usecaseCtx, city, stars, hotelUuid)
	if err != nil {
		return nil, err
	}

	if needSort {
		if isAsc {
			sort.Slice(hotels, func(i, j int) bool {
				return hotels[i].Stars < hotels[j].Stars
			})
		} else {
			sort.Slice(hotels, func(i, j int) bool {
				return hotels[i].Stars > hotels[j].Stars
			})
		}
	}

	return hotels, nil
}
