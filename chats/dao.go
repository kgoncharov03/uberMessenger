package chats

import (
	"context"

	"github.com/joomcode/api/src/misc/generic/timex"
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

	indexOptions := options.Index().SetUnique(true)
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

	for cursor.Next(ctx) {
		var chat *Chat
		if err:=cursor.Decode(&chat); err!=nil {
			return nil, err
		}
		result = append(result, chat)
	}

	return result, nil
}

func (dao *DAO) InitJunk(ctx context.Context) error{
	userID1,err:=primitive.ObjectIDFromHex("5f78829a44202661a33d787a")
	if err!=nil {
		return nil
	}

	userID2,err:=primitive.ObjectIDFromHex("5f78829a44202661a33d787b")
	if err!=nil {
		return nil
	}

	chats:=[]*Chat{
		{
			ID:              primitive.NewObjectID(),
			LastMessageTime: timex.NowMilli(),
			Users: []primitive.ObjectID{
				userID1, userID2,
			},
		},
	}

	if _, err:= dao.collection.InsertOne(ctx, chats[0]); err!=nil {
		return err
	}

	return nil
}

