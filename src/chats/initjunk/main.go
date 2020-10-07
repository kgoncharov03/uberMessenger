package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"uberMessenger/src/chats"
	"uberMessenger/src/common"
	"uberMessenger/src/messages"

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
			LastMessageTime: time.Now().UnixNano(),
			Users: []primitive.ObjectID{
				userID1, userID2,
			},
		},
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: time.Now().UnixNano(),
			Users: []primitive.ObjectID{
				userID1, userID3,
			},
		},
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: time.Now().UnixNano(),
			Users: []primitive.ObjectID{
				userID1, userID2, userID3,
			},
		},
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: time.Now().UnixNano(),
			Users: []primitive.ObjectID{
				userID2, userID3,
			},
		},
	}

	dao.Drop(ctx)



	msgDAO,err := messages.NewDAO(ctx, client)
	if err != nil {
		log.Fatal(err)
	}

	msgDAO.Drop(ctx)

	for i, chat:=range chats {
		dao.AddChat(ctx, chat)
		msg:=&messages.Message{
			ID:             primitive.NewObjectID(),
			From:           chat.Users[0],
			ChatID:         chat.Users[1],
			Text:           "Hello" + strconv.Itoa(i),
			Time:           time.Now().UnixNano(),
			AttachmentLink: nil,
		}

		msgDAO.AddMessage(ctx, msg)
	}
}
