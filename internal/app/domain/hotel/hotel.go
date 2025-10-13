package hotel

import (
	"github.com/google/uuid"
)

type Hotel struct {
	Uuid  uuid.UUID `json:"id" bson:"_id"`
	Name  string    `json:"name" bson:"name"`
	City  string    `json:"city" bson:"city"`
	Stars int32     `json:"stars" bson:"stars"`
}

func NewHotel(name, city string, stars int32) *Hotel {
	return &Hotel{
		Uuid:  uuid.New(),
		Name:  name,
		City:  city,
		Stars: stars,
	}
}
