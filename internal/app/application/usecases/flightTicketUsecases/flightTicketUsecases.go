package flightticketusecases

import (
	"container/heap"
	"context"
	"fmt"
	"sort"
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
	Buy(ctx context.Context, flightTicketUuid uuid.UUID, amountPassengers uint32) (*flightticket.FlightTicket, error)
	GetWithParams(ctx context.Context, flightTicketFilter *flightticket.FlightTicket, cityVia *string, needSort, isAsc bool) ([]Path, []flightticket.FlightTicket, error)
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

func (v *flightTicketUsecase) GetWithParams(ctx context.Context, flightTicketFilter *flightticket.FlightTicket, cityVia *string, needSort, isAsc bool) ([]Path, []flightticket.FlightTicket, error) {
	usecaseCtx, cancel := v.getContext(ctx)
	defer cancel()

	foundFlightTicketsFor2Cities, getErr := v.flightTicketRepo.GetFlightTicketWithParams(usecaseCtx, flightTicketFilter)
	if getErr != nil {
		return nil, nil, getErr
	}

	if cityVia == nil && len(foundFlightTicketsFor2Cities) != 0 {
		if needSort {
			if isAsc {
				sort.Slice(foundFlightTicketsFor2Cities, func(i, j int) bool {
					return *foundFlightTicketsFor2Cities[i].Value < *foundFlightTicketsFor2Cities[j].Value
				})
			} else {
				sort.Slice(foundFlightTicketsFor2Cities, func(i, j int) bool {
					return *foundFlightTicketsFor2Cities[i].Value > *foundFlightTicketsFor2Cities[j].Value
				})
			}
		}
		return nil, foundFlightTicketsFor2Cities, nil
	}

	flightTicketFilterCopy := *flightTicketFilter
	flightTicketFilterCopy.CityFrom = ""
	flightTicketFilterCopy.CityTo = ""

	allFoundFlightTickets, getAllErr := v.flightTicketRepo.GetFlightTicketWithParams(usecaseCtx, &flightTicketFilterCopy)
	if getAllErr != nil {
		return nil, nil, getAllErr
	}

	if len(allFoundFlightTickets) < 2 {
		return nil, nil, nil
	}

	index := make(map[string][]*flightticket.FlightTicket)
	for i := range allFoundFlightTickets {
		f := &allFoundFlightTickets[i]
		index[f.CityFrom] = append(index[f.CityFrom], f)
	}

	for city := range index {
		sort.Slice(index[city], func(i, j int) bool {
			return *index[city][i].TakeOff < *index[city][j].TakeOff
		})
	}

	const maxLayover = int64(24 * 60 * 60)
	pq := &PriorityQueue{}
	heap.Init(pq)

	for _, f := range index[flightTicketFilter.CityFrom] {
		firstDep := *f.TakeOff
		lastArr := f.Arrival
		price := uint64(0)
		if f.Value != nil {
			price = uint64(*f.Value)
		}
		p := &Path{
			Flights:      []*flightticket.FlightTicket{f},
			FirstTakeOff: firstDep,
			LastArrival:  lastArr,
			TotalPrice:   price,
		}
		c := &candidate{
			path:     p,
			priority: p.Duration(),
			priceTie: p.TotalPrice,
		}
		heap.Push(pq, c)
	}

	results := make([]Path, 0, len(allFoundFlightTickets))

	for pq.Len() > 0 {
		item := heap.Pop(pq).(*candidate)
		cur := item.path
		last := cur.Flights[len(cur.Flights)-1]

		if last.CityTo == flightTicketFilter.CityTo {
			if cityVia != nil {
				if wasInViaCity(cur, *cityVia) {
					results = append(results, *cur)
					continue
				}
				continue
			}
			results = append(results, *cur)
			continue
		}

		otherFlights := index[last.CityTo]
		if len(otherFlights) == 0 {
			continue
		}

		for _, flight := range otherFlights {
			if usedInPath(cur, flight.Uuid) {
				continue
			}

			wait := *flight.TakeOff - last.Arrival
			if wait < 0 || wait > maxLayover {
				continue
			}

			newFlights := make([]*flightticket.FlightTicket, len(cur.Flights))
			copy(newFlights, cur.Flights)
			newFlights = append(newFlights, flight)
			newFirst := cur.FirstTakeOff
			newLast := flight.Arrival
			newPrice := cur.TotalPrice
			if flight.Value != nil {
				newPrice += uint64(*flight.Value)
			}
			newPath := &Path{
				Flights:      newFlights,
				FirstTakeOff: newFirst,
				LastArrival:  newLast,
				TotalPrice:   newPrice,
			}
			newCandidate := &candidate{
				path:     newPath,
				priority: newPath.Duration(),
				priceTie: newPath.TotalPrice,
			}
			heap.Push(pq, newCandidate)
		}
	}

	// ToDo: добавить вывод только прямых, без пересадок, путей если такие есть
	// todo: добавить сортировку

	return results, nil, nil
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

func wasInViaCity(path *Path, viaCity string) bool {
	for _, flight := range path.Flights {
		if flight.CityTo == viaCity {
			return true
		}
	}
	return false
}

func usedInPath(p *Path, id uuid.UUID) bool {
	for _, ff := range p.Flights {
		if ff.Uuid == id {
			return true
		}
	}
	return false
}

type Path struct {
	Flights      []*flightticket.FlightTicket
	FirstTakeOff int64
	LastArrival  int64
	TotalPrice   uint64
}

func (p *Path) Duration() int64 {
	if p.FirstTakeOff == 0 {
		return 0
	}
	return p.LastArrival - p.FirstTakeOff
}

type candidate struct {
	path     *Path
	priority int64
	priceTie uint64
	index    int
}

type PriorityQueue []*candidate

func (pq *PriorityQueue) Len() int { return len(*pq) }
func (pq *PriorityQueue) Less(i, j int) bool {
	if (*pq)[i].priority != (*pq)[j].priority {
		return (*pq)[i].priority < (*pq)[j].priority
	}
	return (*pq)[i].priceTie < (*pq)[j].priceTie
}
func (pq *PriorityQueue) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
	(*pq)[i].index = i
	(*pq)[j].index = j
}
func (pq *PriorityQueue) Push(x interface{}) {
	item := x.(*candidate)
	item.index = len(*pq)
	*pq = append(*pq, item)
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1
	*pq = old[0 : n-1]
	return item
}
