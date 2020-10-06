package messages

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const (
	DBName = "messenger"
	CollectionName = "messages"
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
		"chatId": bsonx.Int32(1),
		"time":bsonx.Int32(-1),
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

func (dao *DAO) GetMessagesByChat(ctx context.Context, chatID primitive.ObjectID, limit int, offset int) ([]*Message, error) {
	filter := bson.D{{"chatId", chatID}}
	options:=options.Find()
	options.SetSort(bson.D{{"time", -1}})
	options.SetLimit(int64(limit))
	options.SetSkip(int64(offset))

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var result []*Message

	for cursor.Next(ctx) {
		var message *Message
		if err:=cursor.Decode(&message); err!=nil {
			return nil, err
		}
		result = append(result, message)
	}

	return result, nil
}

func (dao *DAO) AddMessage(ctx context.Context, msg *Message) error {
	if _, err:= dao.collection.InsertOne(ctx, msg); err!=nil {
		return err
	}

	return nil
}

func (dao *DAO) Drop(ctx context.Context) error{
	return dao.collection.Drop(ctx)
}