package flightticketusecases

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/adapters/repository/flightticketrepository"
	"github.com/rom6n/otello/internal/app/domain/flightticket"
)

type FlightTicketUsecases interface {
	Create(ctx context.Context, flightTicket *flightticket.FlightTicket) error
	Update(ctx context.Context, newFlightTicketData *flightticket.FlightTicket) error
	Delete(ctx context.Context, flightTicketUuid uuid.UUID) error
	Get(ctx context.Context, flightTicketUuid uuid.UUID) (*flightticket.FlightTicket, error)
	Buy(ctx context.Context, flightTicketUuid uuid.UUID, amountPassangers uint32) (*flightticket.FlightTicket, error)
	GetWithParams(ctx context.Context, flightTicketFilter *flightticket.FlightTicket, cityVia *string, needSort, isAsc bool) ([]flightticket.FlightTicket, error)
}

type flightTicketUsecase struct {
	flightTicketRepo flightticketrepository.FlightTicketRepository
	timeout          time.Duration
}

func New(flightTicketRepo flightticketrepository.FlightTicketRepository, timeout time.Duration) FlightTicketUsecases {
	return &flightTicketUsecase{
		flightTicketRepo: flightTicketRepo,
		timeout:          timeout,
	}
}

func (v *flightTicketUsecase) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *flightTicketUsecase) Create(ctx context.Context, flightTicket *flightticket.FlightTicket) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.flightTicketRepo.CreateFlightTicket(usecaseCtx, flightTicket)
	if err != nil {
		return err
	}

	return nil
}

func (v *flightTicketUsecase) Update(ctx context.Context, newFlightTicketData *flightticket.FlightTicket) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.flightTicketRepo.UpdateFlightTicket(usecaseCtx, newFlightTicketData)
	if err != nil {
		return err
	}

	return nil
}

func (v *flightTicketUsecase) Delete(ctx context.Context, flightTicketUuid uuid.UUID) error {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	err := v.flightTicketRepo.DeleteFlightTicket(usecaseCtx, flightTicketUuid)
	if err != nil {
		return err
	}

	return nil
}

func (v *flightTicketUsecase) Get(ctx context.Context, flightTicketUuid uuid.UUID) (*flightticket.FlightTicket, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	foundFlightTicket, err := v.flightTicketRepo.GetFlightTicket(usecaseCtx, flightTicketUuid)
	if err != nil {
		return nil, err
	}

	return foundFlightTicket, nil
}

func (v *flightTicketUsecase) GetWithParams(ctx context.Context, flightTicketFilter *flightticket.FlightTicket, cityVia *string, needSort, isAsc bool) ([]flightticket.FlightTicket, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	foundFlightTicketsFor2Cities, getErr := v.flightTicketRepo.GetFlightTicketWithParams(usecaseCtx, flightTicketFilter)
	if getErr != nil {
		return nil, getErr
	}

	fmt.Println(foundFlightTicketsFor2Cities)
	// ToDo: Добавить покупку билетов и поиск по только доступным билетам, добавить поиск по 3 городам. Поиск если нет прямого пути

	return nil, nil
}

func (v *flightTicketUsecase) Buy(ctx context.Context, flightTicketUuid uuid.UUID, amountPassengers uint32) (*flightticket.FlightTicket, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	foundFlightTicket, getErr := v.Get(ctx, flightTicketUuid)
	if getErr != nil {
		return nil, getErr
	}

	if foundFlightTicket.Quantity < amountPassengers {
		return nil, fmt.Errorf("there are no empty seats. only %v left", foundFlightTicket.Quantity)
	}

	buyErr := v.flightTicketRepo.BuyFlightTicket(usecaseCtx, flightTicketUuid, amountPassengers)
	if buyErr != nil {
		return nil, buyErr
	}

	foundFlightTicket.Quantity = amountPassengers
	if foundFlightTicket.Value != nil {
		*foundFlightTicket.Value *= amountPassengers
	}

	return foundFlightTicket, nil
}
