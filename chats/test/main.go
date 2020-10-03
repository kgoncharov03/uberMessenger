package main

import (
	"context"

	"uberMessenger/chats"
	"uberMessenger/common"

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

	dao,err := chats.NewDAO(ctx, client)
	if err != nil {
		panic(err)
	}

	userID,err:=primitive.ObjectIDFromHex("5f78829a44202661a33d787a")
	if err!=nil {
		panic(err)
	}

	chats, err:=dao.GetChatsByUser(ctx, userID)
	if  err!=nil {
		panic(err)
	}

	spew.Dump(chats)
}

