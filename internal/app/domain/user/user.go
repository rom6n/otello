package user

import (
	"github.com/google/uuid"
)

type UserRole string

const (
	RoleAdmin UserRole = "admin"
	RoleUser  UserRole = "user"
)

type User struct {
	Uuid     uuid.UUID `json:"id" bson:"_id"`
	Name     string    `json:"name" bson:"name"`
	Email    string    `json:"email" bson:"email"`
	Password string    `json:"password" bson:"password"`
	Role     UserRole  `json:"role" bson:"role"`
}

func NewUser(name string, email string, password string) *User {
	return &User{
		Uuid:     uuid.New(),
		Name:     name,
		Email:    email,
		Password: password,
		Role:     "user",
	}
}
