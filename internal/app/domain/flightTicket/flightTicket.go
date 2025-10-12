package flightticket

import (
	"github.com/google/uuid"
)

type FlightTicket struct {
	Uuid     uuid.UUID `json:"id" bson:"_id"`
	CityFrom string    `json:"city_from" bson:"city_from"`
	CityTo   string    `json:"city_to" bson:"city_to"`
	Quantity uint32    `json:"quantity" bson:"quantity"`
	Value    *uint32   `json:"value" bson:"value"`
	TakeOff  *int64    `json:"take_off" bson:"take_off"`
	Arrival  int64     `json:"arrival" bson:"arrival"`
}

func NewFlightTicket(cityFrom, cityTo string, quantity uint32, value *uint32, takeOff *int64, arrival int64) *FlightTicket {
	return &FlightTicket{
		Uuid:     uuid.New(),
		CityFrom: cityFrom,
		CityTo:   cityTo,
		Quantity: quantity,
		Value:    value,
		TakeOff:  takeOff,
		Arrival:  arrival,
	}
}
