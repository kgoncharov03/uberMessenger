package storage

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	DBName = "messenger"
	CollectionName = "attachments"
)


type DAO struct {
	client *mongo.Client
	db *mongo.Database
	collection *mongo.Collection
}

func NewDAO(ctx context.Context, client *mongo.Client) (*DAO, error) {
	db := client.Database(DBName)
	collection:=db.Collection(CollectionName)

	return &DAO{
		client:client,
		db:db,
		collection:collection,
	}, nil
}

func (dao *DAO) GetAttachmentByID(ctx context.Context, id primitive.ObjectID) (*Attachment, error) {
	filter := bson.D{{"_id", id}}

	cursor, err := dao.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	var items []*Attachment

	for cursor.Next(ctx) {
		var item *Attachment
		if err:=cursor.Decode(&item); err!=nil {
			return nil, err
		}
		items = append(items, item)
	}

	if len(items) !=1 {
		return nil, errors.New("len(items) !=1")
	}

	return items[0], nil
}

func (dao *DAO) InsertAttachment(ctx context.Context, data *Attachment) error{
	if _, err:= dao.collection.InsertOne(ctx, data); err!=nil {
		return err
	}

	return nil
}


func (dao *DAO) Drop(ctx context.Context) error{
	return dao.collection.Drop(ctx)
}