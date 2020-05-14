package main

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"maibornwolff.de/gomic/mongodb"
	"net/http"
	"strings"
)

func handleFindData(mongodb *mongodb.Mongodb, database string, collection string, writer http.ResponseWriter) {
	documents, err := mongodb.FindData(database, collection, bson.M{"amqpMessage": bson.M{"$exists": true}})
	if err != nil {
		log.Printf("Failed to find data: %s", err)
		return
	}

	log.Printf("Found %d documents", len(documents))

	results := make([]string, len(documents))
	for _, document := range documents {
		results = append(results, fmt.Sprintf("%s", document["amqpMessage"]))
	}

	_, err = writer.Write([]byte(strings.Join(results, "\n")))
	if err != nil {
		log.Printf("Failed to write response: %s", err)
	}
}

func handleAmqpMessage(data []byte, mongo *mongodb.Mongodb, database string, collection string) {
	log.Printf("Inserting AMQP message into MongoDB: %s", string(data))
	_, err := mongo.InsertData(database, collection, bson.M{"amqpMessage": data})
	if err != nil {
		log.Printf("Failed to insert AMQP message into MongoDB: %s", err)
	}
}
