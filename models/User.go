package models

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id"`
	Password  string             `json:"password" bson:"password"`
	Username  string             `json:"username" bson:"username"`
	Email     string             `json:"email" bson:"email"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}
