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
	db         *mongo.Client
	collection *mongo.Collection
}

func NewMongo() (TodoNvm, error) {
	db, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://mongo:123456@127.0.0.1:27017/"))
	collection := db.Database("todos").Collection("todos")
	unique := true
	collection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		bson.D{{"name", 1}},
		&options.IndexOptions{Unique: &unique},
	})
	return &MongoData{db, collection}, err
}

func (M *MongoData) CreateList(list string) error {
	_, err := M.collection.InsertOne(context.TODO(),
		bson.D{{"name", list},
			{"todos", bson.M{}}})
	return err
}

func (M *MongoData) RenameList(list string, name string) error {
	_, err := M.collection.UpdateOne(context.TODO(),
		bson.D{{"name", list}},
		bson.M{"$set": bson.D{{"name", name}}})
	return err
}

func (M *MongoData) Save(list string, todo *Todo) error {
	id := primitive.NewObjectID()
	doc := bson.M{
		"head": todo.Head,
		"desc": todo.Desc,
	}
	_, err := M.collection.UpdateOne(context.TODO(),
		bson.D{{"name", list}},
		bson.M{"$set": bson.M{"todos." + id.Hex(): doc}})
	todo.Id = id.Hex()
	return err
}

func (M *MongoData) Get(list string) ([]Todo, error) {
	res, err := M.collection.Find(context.TODO(), bson.D{{"name", list}})
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
		fmt.Printf("%#v", document)
		list := document["todos"].(primitive.M)
		for id, todo := range list {
			todo := todo.(primitive.M)
			todos = append(todos, Todo{
				id,
				todo["head"].(string),
				todo["desc"].(string),
			})
		}
	}
	return todos, nil
}

func (M *MongoData) Update(list string, todo Todo) error {
	id, err := primitive.ObjectIDFromHex(todo.Id)
	if err != nil {
		return err
	}
	filter := bson.M{
		"name":  bson.M{"$eq": list},
		"todos": bson.M{"$elemMatch": bson.M{"_id": id}},
	}
	_, err = M.collection.
		UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{
				"todos.$.head": todo.Head,
				"todos.$.desc": todo.Desc,
			},
		})
	return err
}

func (M *MongoData) Delete(list string, todo Todo) error {
	id, err := primitive.ObjectIDFromHex(todo.Id)
	if err != nil {
		return err
	}
	_, err = M.collection.UpdateOne(context.TODO(),
		bson.M{"name": list},
		bson.M{"$unset": bson.M{"todos": bson.E{id.String(), ""}}})
	return err
}

func (M *MongoData) Close() error {
	return M.db.Disconnect(context.TODO())
}
