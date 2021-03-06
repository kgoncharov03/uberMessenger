package users

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

const (
	DBName = "messenger"
	CollectionName = "users"
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
		"nickName": bsonx.Int32(1),
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

func (dao *DAO) GetUserByID(ctx context.Context, userID primitive.ObjectID) (*User, error) {
	filter := bson.D{{"_id", userID}}

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var users []*User

	for cursor.Next(ctx) {
		var user *User
		if err:=cursor.Decode(&user); err!=nil {
			return nil, err
		}
		users = append(users, user)
	}

	if len(users) !=1 {
		return nil, errors.New("len(users) !=1")
	}

	return users[0], nil
}

func (dao *DAO) GetUserByNickname(ctx context.Context, nickname string) (*User, error) {
	filter := bson.D{{"nickName", nickname}}

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var users []*User

	for cursor.Next(ctx) {
		var user *User
		if err:=cursor.Decode(&user); err!=nil {
			return nil, err
		}
		users = append(users, user)
	}

	if len(users) !=1 {
		return nil, errors.New("len(users) !=1")
	}

	return users[0], nil
}

func (dao *DAO) InsertUser(ctx context.Context, user *User) error {
	_,err:= dao.collection.InsertOne(ctx, user)
	return err
}

func (dao *DAO) NickNameExists(ctx context.Context, nickname string) (bool, error) {
	filter := bson.D{{"nickName", nickname}}

	cursor, err := dao.collection.Find(context.TODO(), filter)
	if err != nil {
		return false, err
	}

	var users []*User

	for cursor.Next(ctx) {
		var user *User
		if err:=cursor.Decode(&user); err!=nil {
			return false, err
		}
		users = append(users, user)
	}

	if len(users) >0 {
		return true, nil
	}

	return false, err
}

func (dao *DAO) Drop(ctx context.Context) error{
	return dao.collection.Drop(ctx)
}