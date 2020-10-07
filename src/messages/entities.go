package messages

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AttachmentLink struct {
	Type string `bson:"type" json:"type"`
	AttachmentID primitive.ObjectID `bson:"attachmentId" json:"attachmentId"`
	Ext string `bson:"ext" json:"ext"`
	Name string `bson:"name" json:"name"`
}

type Message struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	From primitive.ObjectID `bson:"from" json:"from"`
	ChatID primitive.ObjectID `bson:"chatId" json:"chatId"`
	Text string `bson:"text" json:"text"`
	Time int64 `bson:"time" json:"time"`
	AttachmentLink *AttachmentLink `bson:"attachmentLink,omitempty" json:"attachmentLink,omitempty"`
}
