package storage

import "go.mongodb.org/mongo-driver/bson/primitive"

type Attachment struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	Content []byte `bson:"content" json:"content"`
}
