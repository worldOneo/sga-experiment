package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoData struct {
	db *mongo.Client
}

func NewMongo() (TodoNvm, error) {
	db, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://mongo:123456@127.0.0.1:27017/"))
	return &MongoData{db}, err
}

func (M *MongoData) Save(todo *Todo) error {
	res, err := M.db.Database("todos").Collection("todos").
		InsertOne(context.TODO(), bson.D{
			{"head", todo.Head},
			{"desc", todo.Desc},
		})
	if err != nil {
		return err
	}
	todo.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return nil
}

func (M *MongoData) Get() ([]Todo, error) {
	res, err := M.db.Database("todos").Collection("todos").
		Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	todos := []Todo{}
	var document bson.M
	for res.Next(context.TODO()) {
		err := res.Decode(&document)
		if err != nil {
			return nil, fmt.Errorf("decode: ", err)
		}
		todos = append(todos, Todo{
			document["_id"].(primitive.ObjectID).Hex(),
			document["head"].(string),
			document["desc"].(string),
		})
	}
	return todos, nil
}

func (M *MongoData) Update(todo Todo) error {
	id, err := primitive.ObjectIDFromHex(todo.Id)
	if err != nil {
		return err
	}
	_, err = M.db.Database("todos").Collection("todos").
		UpdateByID(context.TODO(), id, bson.M{
			"$set": bson.D{
				{"head", todo.Head},
				{"desc", todo.Desc},
			},
		})
	return err
}

func (M *MongoData) Delete(todo Todo) error {
	id, err := primitive.ObjectIDFromHex(todo.Id)
	if err != nil {
		return err
	}
	_, err = M.db.Database("todos").Collection("todos").
		DeleteOne(context.TODO(), bson.M{"_id": id})
	return err
}

func (M *MongoData) Close() error {
	return M.db.Disconnect(context.TODO())
}
