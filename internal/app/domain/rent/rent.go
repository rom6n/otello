package rent

import (
	"github.com/google/uuid"
)

// про бронь ничего не сказано. Под вопросом как происходит аренда
type Rent struct {
	Uuid     uuid.UUID `json:"id" bson:"_id"`
	HotelUuid uuid.UUID
	RoomUuid uuid.UUID
	
}

func NewRent(name string, email string, password string) *Rent {
	return &Rent{
		Uuid:     uuid.New(),

	}
}
