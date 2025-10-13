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

	fmt.Println(dbCtx, collection)

	return nil, nil
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
