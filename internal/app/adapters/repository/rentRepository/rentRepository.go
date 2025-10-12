package rentrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/domain/rent"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RentRepository interface {
	CreateRent(ctx context.Context, rent *rent.Rent) error
	GetRentsByHotelRoomUuid(ctx context.Context, hotelRoomUuid uuid.UUID) ([]rent.Rent, error)
	DeleteRent(ctx context.Context, rentUuid uuid.UUID) error
	GetRent(ctx context.Context, rentUuid uuid.UUID) (*rent.Rent, error)
}

type rentRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

func New(dbConnection *mongo.Client, dbName, collectionName string, timeout time.Duration) RentRepository {
	return &rentRepo{
		client:         dbConnection,
		dbName:         dbName,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

func (v *rentRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *rentRepo) getCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *rentRepo) CreateRent(ctx context.Context, rent *rent.Rent) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, err := collection.InsertOne(dbCtx, rent)
	if err != nil {
		return fmt.Errorf("failed to create rent: %v", err)
	}

	return nil
}

func (v *rentRepo) GetRentsByHotelRoomUuid(ctx context.Context, hotelRoomUuid uuid.UUID) ([]rent.Rent, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	cursor, err := collection.Find(dbCtx, bson.D{{Key: "hotel_room_id", Value: hotelRoomUuid}})
	if err != nil {
		return nil, fmt.Errorf("failed to get rents by hotel room uuid: %v", err)
	}

	var rents []rent.Rent
	if err := cursor.All(dbCtx, &rents); err != nil {
		return nil, fmt.Errorf("failed to decode rents: %v", err)
	}

	return rents, nil
}

func (v *rentRepo) GetRent(ctx context.Context, rentUuid uuid.UUID) (*rent.Rent, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var foundedRent rent.Rent

	err := collection.FindOne(dbCtx, bson.D{{Key: "_id", Value: rentUuid}}).Decode(&foundedRent)
	if err != nil {
		return nil, fmt.Errorf("failed to decode rent: %v", err)
	}

	return &foundedRent, nil
}

func (v *rentRepo) DeleteRent(ctx context.Context, rentUuid uuid.UUID) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	isSomethingDeleted, err := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: rentUuid}})
	if err != nil {
		return fmt.Errorf("failed to delete rent: %v", err)
	}

	if isSomethingDeleted.DeletedCount == 0 {
		return fmt.Errorf("failed to delete rent: rent object not found")
	}

	return nil
}
