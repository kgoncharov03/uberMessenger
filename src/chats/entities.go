package chats

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	LastMessageTime time.Time `bson:"lastMessageTime" json:"lastMessageTime"`
	Users []primitive.ObjectID `bson:"users" json:"users"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
}
