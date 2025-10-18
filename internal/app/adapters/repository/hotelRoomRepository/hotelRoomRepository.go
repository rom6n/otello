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
	GetHotelRoomsWithParams(ctx context.Context, filter *hotelroom.FindHotelRoomFilterDTO) ([]hotelroom.HotelRoom, error)
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
			"hotel_uuid":    hotelRoom.HotelUuid,
			"value":         hotelRoom.Value,
			"rooms":         hotelRoom.Rooms,
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

func (v *hotelRoomRepo) GetHotelRoomsWithParams(ctx context.Context, filter *hotelroom.FindHotelRoomFilterDTO) ([]hotelroom.HotelRoom, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	fmt.Printf("filter: %+v\n", filter)

	findParams := ParseParamsToSearchFilter(filter)
	if len(findParams) == 0 {
		findParams = bson.D{}
	}

	fmt.Printf("findParams: %+v\n", findParams)

	cursor, err := collection.Find(dbCtx, findParams)
	if err != nil {
		return nil, fmt.Errorf("failed to find hotel rooms: %v", err)
	}

	var hotelRooms []hotelroom.HotelRoom

	if err := cursor.All(dbCtx, &hotelRooms); err != nil {
		return nil, fmt.Errorf("failed to decode hotel rooms: %v", err)
	}

	return hotelRooms, nil
}

func ParseParamsToSearchFilter(filter *hotelroom.FindHotelRoomFilterDTO) bson.D {
	var findParams bson.D

	if filter.Uuid != uuid.Nil {
		findParams = append(findParams, bson.E{Key: "_id", Value: filter.Uuid})
	}

	if filter.RoomsFrom > 0 || filter.RoomsTo > 0 {
		rangeQuery := bson.D{}

		if filter.RoomsFrom > 0 {
			rangeQuery = append(rangeQuery, bson.E{Key: "$gte", Value: filter.RoomsFrom})
		}
		if filter.RoomsTo > 0 {
			rangeQuery = append(rangeQuery, bson.E{Key: "$lte", Value: filter.RoomsTo})
		}
		if len(rangeQuery) > 0 {
			findParams = append(findParams, bson.E{Key: "rooms", Value: rangeQuery})
		}
	}

	if filter.ValueFrom != nil || filter.ValueTo != nil {
		rangeQuery := bson.D{}

		if filter.ValueFrom != nil {
			rangeQuery = append(rangeQuery, bson.E{Key: "$gte", Value: *filter.ValueFrom})
		}
		if filter.ValueTo != nil {
			rangeQuery = append(rangeQuery, bson.E{Key: "$lte", Value: *filter.ValueTo})
		}
		if len(rangeQuery) > 0 {
			findParams = append(findParams, bson.E{Key: "value", Value: rangeQuery})
		}
	}

	if filter.AmountPeopleFrom > 0 || filter.AmountPeopleTo > 0 {
		rangeQuery := bson.D{}

		if filter.AmountPeopleFrom > 0 {
			rangeQuery = append(rangeQuery, bson.E{Key: "$gte", Value: filter.AmountPeopleFrom})
		}
		if filter.AmountPeopleTo > 0 {
			rangeQuery = append(rangeQuery, bson.E{Key: "$lte", Value: filter.AmountPeopleTo})
		}
		if len(rangeQuery) > 0 {
			findParams = append(findParams, bson.E{Key: "amount_people", Value: rangeQuery})
		}
	}

	if filter.TypeFirst != "" && filter.TypeSecond != "" {
		findParams = append(findParams, bson.E{
			Key:   "type",
			Value: bson.D{{Key: "$in", Value: bson.A{filter.TypeFirst, filter.TypeSecond}}},
		})
	} else if filter.TypeFirst != "" {
		findParams = append(findParams, bson.E{Key: "type", Value: filter.TypeFirst})
	} else if filter.TypeSecond != "" {
		findParams = append(findParams, bson.E{Key: "type", Value: filter.TypeSecond})
	}

	if filter.HotelUuid != uuid.Nil {
		findParams = append(findParams, bson.E{Key: "hotel_uuid", Value: filter.HotelUuid})
	}

	return findParams
}
