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
	IsRented     bool      `json:"is_rented" bson:"is_rented"`
	HotelUuid    uuid.UUID `json:"hotel_uuid" bson:"hotel_uuid"`
	DateFrom     int64     `json:"date_from" bson:"date_from"`
	DateTo       int64     `json:"date_to" bson:"date_to"`
	Days         int32     `json:"days" bson:"days"`
	Rooms        int32     `json:"rooms" bson:"rooms"`
	Type         HotelType `json:"type" bson:"type"`
	AmountPeople int32     `json:"amount_people" bson:"amount_people"`
	Value        int32     `json:"value" bson:"value"`
}

func NewHotelRoom(hotelUuid uuid.UUID, dateFrom, dateTo int64, days, rooms int32, Type HotelType, amountPeople, value int32) *HotelRoom {
	return &HotelRoom{
		Uuid:         uuid.New(),
		IsRented:     false,
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
