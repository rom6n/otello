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

// дата аренды от и до. кол-во дней для проживания. название отеля. кол-во комнат. тип номера standard/large/premium. цена. кол-во человек для проживания. цена
type HotelRoom struct {
	Uuid         uuid.UUID `json:"id" bson:"_id"`
	HotelUuid    uuid.UUID `json:"hotel_uuid" bson:"hotel_uuid"`
	DateFrom     *int64    `json:"date_from" bson:"date_from"`
	DateTo       uint64    `json:"date_to" bson:"date_to"`
	Days         uint32    `json:"days" bson:"days"`
	Rooms        uint32    `json:"rooms" bson:"rooms"`                 // can have 2 values (from, to)
	Type         HotelType `json:"type" bson:"type"`                   // can have 2 values (first, second)
	AmountPeople uint32    `json:"amount_people" bson:"amount_people"` // can have 2 values (from, to)
	Value        *int64    `json:"value" bson:"value"`                 // can have 2 values (from, to)
}

func NewHotelRoom(hotelUuid uuid.UUID, dateFrom *int64, dateTo uint64, days, rooms uint32, Type HotelType, amountPeople uint32, value *int64) *HotelRoom {
	return &HotelRoom{
		Uuid:         uuid.New(),
		HotelUuid:    hotelUuid,
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		Days:         days,
		Rooms:        rooms,
		Type:         Type,
		AmountPeople: amountPeople,
		Value:        value,
	}
}
