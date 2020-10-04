package main

import (
	"context"

	"uberMessenger/src/common"
	"uberMessenger/src/users"

	"github.com/davecgh/go-spew/spew"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	ctx := context.TODO()
	client, err := common.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	dao,err := users.NewDAO(ctx, client)
	if err != nil {
		panic(err)
	}

	userID,err:=primitive.ObjectIDFromHex("5f7a10fc31f3f13dfdc167d7")
	if err!=nil {
		panic(err)
	}

	chats, err:=dao.GetUserByID(ctx, userID)
	if  err!=nil {
		panic(err)
	}

	spew.Dump(chats)
}

