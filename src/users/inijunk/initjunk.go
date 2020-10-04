package main

import (
	"context"
	"log"

	"uberMessenger/src/common"
	"uberMessenger/src/users"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	ctx := context.TODO()
	client, err := common.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	dao,err := users.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}
	dao.Drop(ctx)

	users:=[]*users.User{
		{
			ID:         primitive.NewObjectID(),
			FirstName:  "Gregory",
			SecondName: "Kryloff",
			NickName:   "nagibator",
			Password: "1234",
		},
		{
			ID:         primitive.NewObjectID(),
			FirstName:  "Kirill",
			SecondName: "Goncharov",
			NickName:   "nagibator2",
			Password: "1234",
		},
		{
			ID:         primitive.NewObjectID(),
			FirstName:  "Danil",
			SecondName: "Shaikh",
			NickName:   "my_dick_is_big",
			Password: "1234",
		},
	}

	for _, user:=range users {
		dao.InsertUser(ctx, user)
	}
}