package hotelrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/domain/hotel"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type HotelRepository interface {
	CreateHotel(ctx context.Context, hotel *hotel.Hotel) error
	UpdateHotel(ctx context.Context, hotel *hotel.Hotel) error
	DeleteHotel(ctx context.Context, hotelUuid uuid.UUID) error
	GetHotel(ctx context.Context, hotelUuid uuid.UUID) (*hotel.Hotel, error)
	GetHotelWithParams(ctx context.Context, city string, stars int32, hotelUuid uuid.UUID, starsFrom, starsTo uint32) ([]hotel.Hotel, error)
}

type hotelRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

func New(dbConnection *mongo.Client, dbName, collectionName string, timeout time.Duration) HotelRepository {
	return &hotelRepo{
		client:         dbConnection,
		dbName:         dbName,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

func (v *hotelRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *hotelRepo) getCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *hotelRepo) CreateHotel(ctx context.Context, hotel *hotel.Hotel) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, err := collection.InsertOne(dbCtx, hotel)
	if err != nil {
		return fmt.Errorf("failed to create hotel: %v", err)
	}

	return nil
}

func (v *hotelRepo) UpdateHotel(ctx context.Context, hotel *hotel.Hotel) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	update := bson.M{
		"$set": bson.M{
			"name":  hotel.Name,
			"city":  hotel.City,
			"stars": hotel.Stars,
		},
	}

	_, err := collection.UpdateByID(dbCtx, hotel.Uuid, update)
	if err != nil {
		return fmt.Errorf("failed to update hotel: %v", err)
	}

	return nil
}

func (v *hotelRepo) DeleteHotel(ctx context.Context, hotelUuid uuid.UUID) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	result, err := collection.DeleteOne(dbCtx, bson.D{{Key: "_id", Value: hotelUuid}})
	if err != nil {
		return fmt.Errorf("failed to delete hotel: %v", err)
	}
	if result.DeletedCount < 1 {
		return fmt.Errorf("failed to delete hotel: hotel not found")
	}

	return nil
}

func (v *hotelRepo) GetHotel(ctx context.Context, hotelUuid uuid.UUID) (*hotel.Hotel, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var foundedHotel hotel.Hotel

	err := collection.FindOne(dbCtx, bson.D{{Key: "_id", Value: hotelUuid}}).Decode(&foundedHotel)
	if err != nil {
		return nil, fmt.Errorf("failed to find hotel: %v", err)
	}

	return &foundedHotel, nil
}

func (v *hotelRepo) GetHotelWithParams(ctx context.Context, city string, stars int32, hotelUuid uuid.UUID, starsFrom, starsTo uint32) ([]hotel.Hotel, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	findParams := bson.D{}

	if city != "" {
		findParams = append(findParams, bson.E{Key: "city", Value: city})
	}

	if stars > 0 {
		findParams = append(findParams, bson.E{Key: "stars", Value: stars})
	}

	if starsFrom > 0 {
		rangeQuery := bson.D{}
		rangeQuery = append(rangeQuery, bson.E{Key: "$gte", Value: starsFrom})
		findParams = append(findParams, bson.E{Key: "stars", Value: rangeQuery})
	}

	if starsTo > 0 {
		rangeQuery := bson.D{}
		rangeQuery = append(rangeQuery, bson.E{Key: "$lte", Value: starsTo})
		findParams = append(findParams, bson.E{Key: "stars", Value: rangeQuery})
	}

	if hotelUuid != uuid.Nil {
		findParams = append(findParams, bson.E{Key: "_id", Value: hotelUuid})
	}

	cursor, err := collection.Find(dbCtx, findParams)
	if err != nil {
		return nil, fmt.Errorf("failed to find hotels: %v", err)
	}

	var hotels []hotel.Hotel

	if err := cursor.All(dbCtx, &hotels); err != nil {
		return nil, fmt.Errorf("failed to decode hotels: %v", err)
	}

	return hotels, nil
}
