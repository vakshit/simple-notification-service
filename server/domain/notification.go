package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Notification struct {
	ID      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Time    string             `json:"time" bson:"time"`
	Message string             `json:"message" bson:"message"`
	Email   string             `json:"email" bson:"email"`
}
