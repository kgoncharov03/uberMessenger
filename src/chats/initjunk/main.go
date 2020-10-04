package main

import (
	"context"
	"log"
	"time"

	"uberMessenger/src/chats"
	"uberMessenger/src/common"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	ctx := context.TODO()
	client, err := common.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	dao,err := chats.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	userID1,err:=primitive.ObjectIDFromHex("5f7a10fc31f3f13dfdc167d6")
	if err!=nil {
		panic(err)
	}

	userID2,err:=primitive.ObjectIDFromHex("5f7a10fc31f3f13dfdc167d7")
	if err!=nil {
		panic(err)
	}

	userID3,err:=primitive.ObjectIDFromHex("5f7a10fc31f3f13dfdc167d8")
	if err!=nil {
		panic(err)
	}

	chats:=[]*chats.Chat{
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: time.Now(),
			Users: []primitive.ObjectID{
				userID1, userID2,
			},
		},
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: time.Now(),
			Users: []primitive.ObjectID{
				userID1, userID3,
			},
		},
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: time.Now(),
			Users: []primitive.ObjectID{
				userID1, userID2, userID3,
			},
		},
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: time.Now(),
			Users: []primitive.ObjectID{
				userID2, userID3,
			},
		},
	}

	dao.Drop(ctx)

	for _, chat:=range chats {
		dao.AddChat(ctx, chat)
	}

}
