package rent

import (
	"github.com/google/uuid"
)

type Rent struct {
	Uuid       uuid.UUID `json:"id" bson:"_id"`
	RoomUuid   uuid.UUID `json:"hotel_room_id" bson:"hotel_room_id"`
	RenterUuid uuid.UUID `json:"renter_id" bson:"renter_id"`
	DateFrom   int64     `json:"date_from" bson:"date_from"`
	DateTo     int64     `json:"date_to" bson:"date_to"`
}

func NewRent(roomUuid, renterUuid uuid.UUID, dateFrom int64, dateTo int64) *Rent {
	return &Rent{
		Uuid:       uuid.New(),
		RoomUuid:   roomUuid,
		RenterUuid: renterUuid,
		DateFrom:   dateFrom,
		DateTo:     dateTo,
	}
}
