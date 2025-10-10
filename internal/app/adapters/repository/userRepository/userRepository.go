package userrepository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rom6n/otello/internal/app/domain/user"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *user.User) error
	GetUser(ctx context.Context, email string) (*user.User, error)
	UpdateUserName(ctx context.Context, userId uuid.UUID, newName string) error
}

type userRepo struct {
	client         *mongo.Client
	dbName         string
	collectionName string
	timeout        time.Duration
}

func New(dbConnection *mongo.Client, dbName, collectionName string, timeout time.Duration) UserRepository {
	return &userRepo{
		client:         dbConnection,
		dbName:         dbName,
		collectionName: collectionName,
		timeout:        timeout,
	}
}

func (v *userRepo) getContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, v.timeout)
}

func (v *userRepo) getCollection() *mongo.Collection {
	return v.client.Database(v.dbName).Collection(v.collectionName)
}

func (v *userRepo) CreateUser(ctx context.Context, user *user.User) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	_, err := collection.InsertOne(dbCtx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

func (v *userRepo) GetUser(ctx context.Context, email string) (*user.User, error) {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	var user user.User
	if err := collection.FindOne(dbCtx, bson.D{{Key: "email", Value: email}}).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to get user from database: %v", err)
	}

	return &user, nil
}

func (v *userRepo) UpdateUserName(ctx context.Context, userId uuid.UUID, newName string) error {
	dbCtx, cancel := v.getContext(ctx)
	defer cancel()

	collection := v.getCollection()

	update := bson.M{
		"$set": bson.M{
			"name": newName,
		},
	}

	if _, err := collection.UpdateByID(dbCtx, userId, update); err != nil {
		return fmt.Errorf("failed to update name: %v", err)
	}

	return nil
}
