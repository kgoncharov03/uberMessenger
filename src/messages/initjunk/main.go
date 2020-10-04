package main

import (
	"context"
	"time"

	"uberMessenger/src/common"
	"uberMessenger/src/messages"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	ctx := context.TODO()
	client, err := common.NewClient()
	if err != nil {
		panic(err)
	}
	defer client.Disconnect(ctx)

	dao,err := messages.NewDAO(ctx, client)
	if err != nil {
		panic(err)
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

	chatID1,err:=primitive.ObjectIDFromHex("5f7a1715fd17b329f885e7bc")
	if err!=nil {
		panic(err)
	}

	chatID2,err:=primitive.ObjectIDFromHex("5f7a1715fd17b329f885e7bd")
	if err!=nil {
		panic(err)
	}

	chatID3,err:=primitive.ObjectIDFromHex("5f7a1715fd17b329f885e7be")
	if err!=nil {
		panic(err)
	}

	chatID4,err:=primitive.ObjectIDFromHex("5f7a1715fd17b329f885e7be")
	if err!=nil {
		panic(err)
	}


	msgs:=[]*messages.Message{
		{
			ID:     primitive.NewObjectID(),
			From:   userID1,
			ChatID: chatID1,
			Text:   "ПИСЯ",
			Time:   time.Now(),
		},
		{
			ID:     primitive.NewObjectID(),
			From:   userID2,
			ChatID: chatID1,
			Text:   "СИСЯ",
			Time:   time.Now().Add(time.Hour),
		},

		{
			ID:     primitive.NewObjectID(),
			From:   userID1,
			ChatID: chatID2,
			Text:   "ХУЙ",
			Time:   time.Now(),
		},
		{
			ID:     primitive.NewObjectID(),
			From:   userID3,
			ChatID: chatID2,
			Text:   "ПИЗДА",
			Time:   time.Now().Add(time.Hour),
		},

		{
			ID:     primitive.NewObjectID(),
			From:   userID2,
			ChatID: chatID3,
			Text:   "АААА",
			Time:   time.Now(),
		},
		{
			ID:     primitive.NewObjectID(),
			From:   userID3,
			ChatID: chatID3,
			Text:   "ББББ",
			Time:   time.Now().Add(time.Hour),
		},

		{
			ID:     primitive.NewObjectID(),
			From:   userID2,
			ChatID: chatID4,
			Text:   "ЫЫЫ",
			Time:   time.Now(),
		},
		{
			ID:     primitive.NewObjectID(),
			From:   userID3,
			ChatID: chatID4,
			Text:   "ФФФФ",
			Time:   time.Now().Add(time.Hour),
		},
		{
			ID:     primitive.NewObjectID(),
			From:   userID1,
			ChatID: chatID4,
			Text:   "АЫЫЫЫ",
			Time:   time.Now().Add(time.Hour*2),
		},
	}


	for _, msg:=range msgs {
		dao.AddMessage(ctx, msg)
	}
}
