package chats

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	LastMessageTime int64 `bson:"-" json:"lastMessageTime,omitempty"`
	Users []primitive.ObjectID `bson:"users" json:"users"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
	LastMessage string `bson:"-" json:"lastMessage,omitempty"`
}
