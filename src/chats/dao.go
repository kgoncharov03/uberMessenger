package chats

import (
	"context"
	"errors"

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


func (dao *DAO) GetChatByID(ctx context.Context, chatID primitive.ObjectID) (*Chat, error) {
	filter := bson.D{{"_id", chatID}}

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var chats []*Chat

	for cursor.Next(ctx) {
		var chat *Chat
		if err:=cursor.Decode(&chat); err!=nil {
			return nil, err
		}
		chats = append(chats, chat)
	}

	if len(chats) !=1 {
		return nil, errors.New("len(chats) !=1")
	}

	return chats[0], nil
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
