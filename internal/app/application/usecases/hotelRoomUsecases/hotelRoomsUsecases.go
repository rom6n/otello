package hotelroomusecases

import (
	"context"
	"time"

	"github.com/google/uuid"
	hotelroomrepository "github.com/rom6n/otello/internal/app/adapters/repository/hotelroomrepository"
	"github.com/rom6n/otello/internal/app/domain/hotelroom"
)

type HotelRoomUsecases interface {
	Create(ctx context.Context, hotelRoom *hotelroom.HotelRoom) error
	Update(ctx context.Context, newHotelRoomData *hotelroom.HotelRoom) error
	Delete(ctx context.Context, hotelRoomUuid uuid.UUID) error
	Get(ctx context.Context, hotelRoomUuid uuid.UUID) (*hotelroom.HotelRoom, error)
}

type hotelRoomUsecase struct {
	hotelRoomRepo hotelroomrepository.HotelRoomRepository
	timeout       time.Duration
}

func New(hotelRoomRepo hotelroomrepository.HotelRoomRepository, timeout time.Duration) HotelRoomUsecases {
	return &hotelRoomUsecase{
		hotelRoomRepo: hotelRoomRepo,
		timeout:       timeout,
	}
}

func (v *hotelRoomUsecase) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *hotelRoomUsecase) Create(ctx context.Context, hotelRoom *hotelroom.HotelRoom) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.hotelRoomRepo.CreateHotelRoom(usecaseCtx, hotelRoom)
	if err != nil {
		return err
	}

	return nil
}

func (v *hotelRoomUsecase) Update(ctx context.Context, newHotelRoomData *hotelroom.HotelRoom) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.hotelRoomRepo.UpdateHotelRoom(usecaseCtx, newHotelRoomData)
	if err != nil {
		return err
	}

	return nil
}

func (v *hotelRoomUsecase) Delete(ctx context.Context, hotelRoomUuid uuid.UUID) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.hotelRoomRepo.DeleteHotelRoom(usecaseCtx, hotelRoomUuid)
	if err != nil {
		return err
	}

	return nil
}

func (v *hotelRoomUsecase) Get(ctx context.Context, hotelRoomUuid uuid.UUID) (*hotelroom.HotelRoom, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	hotel, err := v.hotelRoomRepo.GetHotelRoom(usecaseCtx, hotelRoomUuid)
	if err != nil {
		return nil, err
	}

	return hotel, nil
}
