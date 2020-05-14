package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type Mongodb struct {
	client *mongo.Client
}

func Connect(mongodbHost string) *Mongodb {
	client, err := mongo.NewClient(options.Client().ApplyURI(mongodbHost))
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %s", err)
	}

	err = client.Connect(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %s", err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %s", err)
	}

	log.Printf("Connected to MongoDB")
	return &Mongodb{
		client: client,
	}
}

func (mongodb *Mongodb) Shutdown() {
	err := mongodb.client.Disconnect(context.Background())
	if err != nil {
		log.Printf("Failed to disconnect from MongoDB: %s", err)
	}
}

func (mongodb *Mongodb) InsertData(database string, collection string, data interface{}) (*mongo.InsertOneResult, error) {
	return mongodb.client.Database(database).Collection(collection).InsertOne(context.Background(), data)
}

func (mongodb *Mongodb) FindData(database string, collection string, selector bson.M) ([]bson.M, error) {
	cursor, err := mongodb.client.Database(database).Collection(collection).Find(context.Background(), selector)
	if err != nil {
		return nil, err
	}

	defer cursor.Close(context.Background())

	results := make([]bson.M, 1)
	for cursor.Next(context.Background()) {
		var result bson.M
		err = cursor.Decode(&result)
		if err != nil {
			log.Printf("Failed to decode result: %s", err)
		}
		results = append(results, result)
	}

	return results, nil
}
