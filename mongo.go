package todo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoData struct {
	db             *mongo.Client
	todoCollection *mongo.Collection
}

func NewMongo() (TodoNvm, error) {
	db, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://mongo:123456@127.0.0.1:27017/"))
	if err != nil {
		return nil, err
	}
	todoCollection := db.Database("todos").Collection("todos")
	_, err = todoCollection.Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		bson.D{{"list", 1}},
		&options.IndexOptions{},
	})
	return &MongoData{db, todoCollection}, err
}

func (M *MongoData) CreateList(list string) error {
	return nil
}

func (M *MongoData) RenameList(list string, name string) error {
	_, err := M.todoCollection.UpdateMany(context.TODO(),
		bson.D{{"list", list}},
		bson.M{"$set": bson.D{{"list", name}}})
	return err
}

func (M *MongoData) Save(list string, todo *Todo) error {
	id := primitive.NewObjectID()
	doc := bson.M{
		"list": list,
		"head": todo.Head,
		"desc": todo.Desc,
	}
	_, err := M.todoCollection.InsertOne(context.TODO(), doc)
	todo.Id = id.Hex()
	return err
}

func (M *MongoData) Get(list string) ([]Todo, error) {
	res, err := M.todoCollection.Find(context.TODO(), bson.D{{"list", list}})
	if err != nil {
		return nil, err
	}
	todos := []Todo{}
	var document bson.M
	defer res.Close(context.TODO())
	for res.Next(context.TODO()) {
		err := res.Decode(&document)
		if err != nil {
			return nil, fmt.Errorf("decode: %v", err)
		}
		todos = append(todos, Todo{
			document["_id"].(primitive.ObjectID).Hex(),
			document["head"].(string),
			document["desc"].(string),
		})
	}
	return todos, nil
}

func (M *MongoData) Update(list string, todo Todo) error {
	id, err := primitive.ObjectIDFromHex(todo.Id)
	if err != nil {
		return err
	}
	filter := bson.M{
		"_id": bson.M{"$eq": id},
	}
	_, err = M.todoCollection.
		UpdateOne(context.TODO(), filter, bson.M{
			"$set": bson.M{
				"head": todo.Head,
				"desc": todo.Desc,
			},
		})
	return err
}

func (M *MongoData) Delete(list string, todo Todo) error {
	id, err := primitive.ObjectIDFromHex(todo.Id)
	if err != nil {
		return err
	}
	_, err = M.todoCollection.DeleteOne(context.TODO(),	bson.M{"_id": bson.M{"$eq": id}})
	return err
}

func (M *MongoData) Close() error {
	return M.db.Disconnect(context.TODO())
}
