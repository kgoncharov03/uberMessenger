package users

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID primitive.ObjectID `bson:"_id" json:"id"`
	FirstName string`bson:"firstName" json:"firstName"`
	SecondName string `bson:"secondName" json:"secondName"`
	NickName string `bson:"nickName" json:"nickName"`
}
