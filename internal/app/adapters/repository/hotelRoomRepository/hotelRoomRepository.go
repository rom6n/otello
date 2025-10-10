package hotelRoomroomrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/domain/hotelroom"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type HotelRoomRepository interface {
	CreateHotelRoom(ctx context.Context, hotelRoom *hotelroom.HotelRoom) error
	UpdateHotelRoom(ctx context.Context, hotelRoom *hotelroom.HotelRoom) error
	DeleteHotelRoom(ctx context.Context, hotelRoomUuid uuid.UUID) error
	GetHotelRoom(ctx context.Context, hotelRoomUuid uuid.UUID) (*hotelroom.HotelRoom, error)
}

type hotelRoomRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

func New(dbConnection *mongo.Client, dbName, collectionName string, timeout time.Duration) HotelRoomRepository {
	return &hotelRoomRepo{
		client:         dbConnection,
		dbName:         dbName,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

func (v *hotelRoomRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *hotelRoomRepo) getCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *hotelRoomRepo) CreateHotelRoom(ctx context.Context, hotelRoom *hotelroom.HotelRoom) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, err := collection.InsertOne(dbCtx, hotelRoom)
	if err != nil {
		return fmt.Errorf("failed to create hotel room: %v", err)
	}

	return nil
}

func (v *hotelRoomRepo) UpdateHotelRoom(ctx context.Context, hotelRoom *hotelroom.HotelRoom) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	update := bson.M{
		"$set": bson.M{
			"date_from":     hotelRoom.DateFrom,
			"date_to":       hotelRoom.DateTo,
			"hotel_uuid":    hotelRoom.HotelUuid,
			"is_rented":     hotelRoom.IsRented,
			"value":         hotelRoom.Value,
			"rooms":         hotelRoom.Rooms,
			"days":          hotelRoom.Days,
			"type":          hotelRoom.Type,
			"amount_people": hotelRoom.AmountPeople,
		},
	}

	_, err := collection.UpdateByID(dbCtx, hotelRoom.Uuid, update)
	if err != nil {
		return fmt.Errorf("failed to update hotel room: %v", err)
	}

	return nil
}

func (v *hotelRoomRepo) DeleteHotelRoom(ctx context.Context, hotelRoomUuid uuid.UUID) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, err := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: hotelRoomUuid}})
	if err != nil {
		return fmt.Errorf("failed to delete hotel room: %v", err)
	}

	return nil
}

func (v *hotelRoomRepo) GetHotelRoom(ctx context.Context, hotelRoomUuid uuid.UUID) (*hotelroom.HotelRoom, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var hotelRoom hotelroom.HotelRoom

	err := collection.FindOne(dbCtx, bson.D{{Key: "_id", Value: hotelRoomUuid}}).Decode(&hotelRoom)
	if err != nil {
		return nil, fmt.Errorf("failed to find hotel room: %v", err)
	}

	return &hotelRoom, nil
}
