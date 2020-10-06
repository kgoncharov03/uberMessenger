package messages

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	From primitive.ObjectID `bson:"from" json:"from"`
	ChatID primitive.ObjectID `bson:"chatId" json:"chatId"`
	Text string `bson:"text" json:"text"`
	Time int64 `bson:"time" json:"time"`
}
