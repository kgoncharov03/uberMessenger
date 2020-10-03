package chats

import (
	"github.com/joomcode/api/src/misc/generic/timex"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Chat struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	LastMessageTime timex.TimeMilli `bson:"lastMessageTime" json:"lastMessageTime"`
	Users []primitive.ObjectID `bson:"users" json:"users"`
	Name string `bson:"name,omitempty" json:"name,omitempty"`
}
