package hotelroom

import (
	"github.com/google/uuid"
)

type HotelType string

const (
	Standard HotelType = "standard"
	Lagre    HotelType = "large"
	Premium  HotelType = "premium"
)

type HotelRoom struct {
	Uuid         uuid.UUID `json:"id" bson:"_id"`
	HotelUuid    uuid.UUID `json:"hotel_uuid" bson:"hotel_uuid"`
	Rooms        uint32    `json:"rooms" bson:"rooms"`                 // can have 2 values (from, to)
	Type         HotelType `json:"type" bson:"type"`                   // can have 2 values (first, second)
	AmountPeople uint32    `json:"amount_people" bson:"amount_people"` // can have 2 values (from, to)
	Value        *int64    `json:"value" bson:"value"`                 // can have 2 values (from, to)
}

type FindHotelRoomDTO struct {
	Date         *int64
	Rooms        uint32
	Type         HotelType
	AmountPeople uint32
	Value        *int64
}
type FindHotelRoomFilterDTO struct {
	Uuid             uuid.UUID
	HotelUuid        uuid.UUID
	DateFrom         *int64
	DateTo           *int64
	RoomsFrom        uint32
	RoomsTo          uint32
	TypeFirst        string
	TypeSecond       string
	AmountPeopleFrom uint32
	AmountPeopleTo   uint32
	ValueFrom        *int64
	ValueTo          *int64
	Arrange          string
}

func NewHotelRoom(hotelUuid uuid.UUID, rooms uint32, Type HotelType, amountPeople uint32, value *int64) *HotelRoom {
	return &HotelRoom{
		Uuid:         uuid.New(),
		HotelUuid:    hotelUuid,
		Rooms:        rooms,
		Type:         Type,
		AmountPeople: amountPeople,
		Value:        value,
	}
}
