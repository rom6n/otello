package hotelroomusecases

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"
	hotelroomrepository "github.com/rom6n/otello/internal/app/adapters/repository/hotelroomrepository"
	"github.com/rom6n/otello/internal/app/adapters/repository/rentrepository"
	"github.com/rom6n/otello/internal/app/domain/hotelroom"
)

type HotelRoomUsecases interface {
	Create(ctx context.Context, hotelRoom *hotelroom.HotelRoom) error
	Update(ctx context.Context, newHotelRoomData *hotelroom.HotelRoom) error
	Delete(ctx context.Context, hotelRoomUuid uuid.UUID) error
	Get(ctx context.Context, hotelRoomUuid uuid.UUID) (*hotelroom.HotelRoom, error)
	GetWithParams(ctx context.Context, filter *hotelroom.FindHotelRoomFilterDTO) ([]hotelroom.HotelRoom, error)
}

type hotelRoomUsecase struct {
	hotelRoomRepo hotelroomrepository.HotelRoomRepository
	rentRepo      rentrepository.RentRepository
	timeout       time.Duration
}

func New(hotelRoomRepo hotelroomrepository.HotelRoomRepository, rentRepo rentrepository.RentRepository, timeout time.Duration) HotelRoomUsecases {
	return &hotelRoomUsecase{
		hotelRoomRepo: hotelRoomRepo,
		rentRepo:      rentRepo,
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

func (v *hotelRoomUsecase) GetWithParams(ctx context.Context, filter *hotelroom.FindHotelRoomFilterDTO) ([]hotelroom.HotelRoom, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	hotelRooms, err := v.hotelRoomRepo.GetHotelRoomsWithParams(usecaseCtx, filter)
	if err != nil {
		return nil, err
	}

	var availableHotelRooms []hotelroom.HotelRoom
	if filter.DateFrom != nil {
		checkIn := *filter.DateFrom
		checkOut := *filter.DateTo

		for _, hotelRoom := range hotelRooms {
			rents, getRentsErr := v.rentRepo.GetRentsByHotelRoomUuid(usecaseCtx, hotelRoom.Uuid)
			if getRentsErr != nil {
				return nil, getRentsErr
			}

			isAvailable := true

			for _, rent := range rents {
				if rent.DateFrom < checkOut && rent.DateTo > checkIn {
					isAvailable = false
					break
				}
			}

			if isAvailable {
				availableHotelRooms = append(availableHotelRooms, hotelRoom)
			}
		}
	} else {
		availableHotelRooms = hotelRooms
	}

	if filter.Arrange != "" {
		if filter.Arrange == "asc" {
			sort.Slice(availableHotelRooms, func(i, j int) bool {
				return *availableHotelRooms[i].Value < *availableHotelRooms[j].Value
			})
		} else {
			sort.Slice(availableHotelRooms, func(i, j int) bool {
				return *availableHotelRooms[i].Value > *availableHotelRooms[j].Value
			})
		}
	}

	return availableHotelRooms, nil
}
