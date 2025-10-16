package flightticketrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/domain/flightticket"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type FlightTicketRepository interface {
	CreateFlightTicket(ctx context.Context, flightTicket *flightticket.FlightTicket) error
	UpdateFlightTicket(ctx context.Context, flightTicket *flightticket.FlightTicket) error
	DeleteFlightTicket(ctx context.Context, flightTicketUuid uuid.UUID) error
	GetFlightTicket(ctx context.Context, flightTicketUuid uuid.UUID) (*flightticket.FlightTicket, error)
	BuyFlightTicket(ctx context.Context, flightTicketUuid uuid.UUID, amountPassengers uint32) error
	GetFlightTicketWithParams(ctx context.Context, flightTicketFilter *flightticket.FlightTicket) ([]flightticket.FlightTicket, error)
}

type flightTicketRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

func New(dbConnection *mongo.Client, dbName, collectionName string, timeout time.Duration) FlightTicketRepository {
	return &flightTicketRepo{
		client:         dbConnection,
		dbName:         dbName,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

func (v *flightTicketRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *flightTicketRepo) getCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *flightTicketRepo) CreateFlightTicket(ctx context.Context, flightTicket *flightticket.FlightTicket) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, err := collection.InsertOne(dbCtx, flightTicket)
	if err != nil {
		return fmt.Errorf("failed to create flight ticket: %v", err)
	}

	return nil
}

func (v *flightTicketRepo) UpdateFlightTicket(ctx context.Context, flightTicket *flightticket.FlightTicket) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	update := bson.M{
		"$set": bson.M{
			"city_from": flightTicket.CityFrom,
			"city_to":   flightTicket.CityTo,
			"quantity":  flightTicket.Quantity,
			"value":     flightTicket.Value,
			"take_off":  flightTicket.TakeOff,
			"arrival":   flightTicket.Arrival,
		},
	}

	_, err := collection.UpdateByID(dbCtx, flightTicket.Uuid, update)
	if err != nil {
		return fmt.Errorf("failed to update flight ticket: %v", err)
	}

	return nil
}

func (v *flightTicketRepo) DeleteFlightTicket(ctx context.Context, flightTicketUuid uuid.UUID) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, err := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: flightTicketUuid}})
	if err != nil {
		return fmt.Errorf("failed to delete flight ticket: %v", err)
	}

	return nil
}

func (v *flightTicketRepo) GetFlightTicket(ctx context.Context, flightTicketUuid uuid.UUID) (*flightticket.FlightTicket, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var foundedFlightTicket flightticket.FlightTicket

	err := collection.FindOne(dbCtx, bson.D{{Key: "_id", Value: flightTicketUuid}}).Decode(&foundedFlightTicket)
	if err != nil {
		return nil, fmt.Errorf("failed to find flight ticket: %v", err)
	}

	return &foundedFlightTicket, nil
}

func (v *flightTicketRepo) GetFlightTicketWithParams(ctx context.Context, flightTicketFilter *flightticket.FlightTicket) ([]flightticket.FlightTicket, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	findParams := ParseParamsToSearchFilter(flightTicketFilter)
	if len(findParams) == 0 {
		findParams = bson.D{}
	}

	cursor, findErr := collection.Find(dbCtx, findParams)
	if findErr != nil {
		return nil, fmt.Errorf("failed to find flight ticket: %v", findErr)
	}

	var foundedFlightTickets []flightticket.FlightTicket

	if decodeErr := cursor.All(dbCtx, &foundedFlightTickets); decodeErr != nil {
		return nil, fmt.Errorf("failed to decode flight tickets: %v", decodeErr)
	}

	return foundedFlightTickets, nil
}

func (v *flightTicketRepo) BuyFlightTicket(ctx context.Context, flightTicketUuid uuid.UUID, amountPassengers uint32) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	update := bson.M{"$inc": bson.M{"quantity": -int64(amountPassengers)}}

	isFoundAndUpdated, err := collection.UpdateByID(dbCtx, flightTicketUuid, update)
	if err != nil {
		return fmt.Errorf("failed to find flight ticket: %v", err)
	}

	if isFoundAndUpdated.ModifiedCount == 0 {
		return fmt.Errorf("flight ticket not exists")
	}

	return nil
}

func ParseParamsToSearchFilter(filter *flightticket.FlightTicket) bson.D {
	var findParams bson.D

	if filter.Uuid != uuid.Nil {
		findParams = append(findParams, bson.E{Key: "_id", Value: filter.Uuid})
	}

	var amountQuantity uint32 = 1
	if filter.Quantity > 0 {
		amountQuantity = filter.Quantity
	}
	rangeQuery := bson.D{}
	rangeQuery = append(rangeQuery, bson.E{Key: "$gte", Value: amountQuantity})
	findParams = append(findParams, bson.E{Key: "quantity", Value: rangeQuery})

	if filter.CityFrom != "" {
		findParams = append(findParams, bson.E{Key: "city_from", Value: filter.CityFrom})
		findParams = append(findParams, bson.E{Key: "city_to", Value: filter.CityTo})

	}

	if filter.TakeOff != nil {
		rangeQuery1 := bson.D{}
		rangeQuery2 := bson.D{}

		rangeQuery1 = append(rangeQuery1, bson.E{Key: "$gte", Value: filter.TakeOff})

		rangeQuery2 = append(rangeQuery2, bson.E{Key: "$lte", Value: filter.Arrival})

		findParams = append(findParams, bson.E{Key: "take_off", Value: rangeQuery1})
		findParams = append(findParams, bson.E{Key: "take_off", Value: rangeQuery2})
	}

	return findParams
}
