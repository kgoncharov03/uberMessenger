package chats

import (
	"context"
	"time"

	"github.com/davecgh/go-spew/spew"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const (
	DBName = "messenger"
	CollectionName = "chats"
)


type DAO struct {
	client *mongo.Client
	db *mongo.Database
	collection *mongo.Collection
}

func NewDAO(ctx context.Context, client *mongo.Client) (*DAO, error) {
	db := client.Database(DBName)
	collection:=db.Collection(CollectionName)

	indexOptions := options.Index().SetUnique(false)
	indexKeys := bsonx.MDoc{
		"users": bsonx.Int32(1),
	}

	noteIndexModel := mongo.IndexModel{
		Options: indexOptions,
		Keys:    indexKeys,
	}

	_, err := collection.Indexes().CreateOne(ctx, noteIndexModel)
	if err != nil {
		return nil, err
	}

	return &DAO{
		client:client,
		db:db,
		collection:collection,
	}, nil
}

func (dao *DAO) GetChatsByUser(ctx context.Context, userID primitive.ObjectID) ([]*Chat, error) {
	filter := bson.D{{"users", userID}}

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var result []*Chat

	spew.Dump(filter)

	for cursor.Next(ctx) {
		var chat *Chat
		if err:=cursor.Decode(&chat); err!=nil {
			return nil, err
		}
		result = append(result, chat)
	}

	return result, nil
}

func (dao *DAO) AddChat(ctx context.Context, chat *Chat) error {
	if _, err:= dao.collection.InsertOne(ctx, chat); err!=nil {
		return err
	}

	return nil
}

func (dao *DAO) Drop(ctx context.Context) error{
	return dao.collection.Drop(ctx)
}

func (dao *DAO) InitJunk(ctx context.Context) error{
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

	chats:=[]*Chat{
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

	return nil
}

