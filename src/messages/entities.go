package messages

import (
	"github.com/joomcode/api/src/misc/generic/timex"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Message struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	From primitive.ObjectID `bson:"from" json:"from"`
	ChatID primitive.ObjectID `bson:"chatId" json:"chatId"`
	Text string `bson:"text" json:"text"`
	Time timex.TimeMilli `bson:"time" json:"time"`
}
