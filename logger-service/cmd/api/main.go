package main

import (
	"log"

	"log-service/data"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const mongoURL = "mongodb://mongo:27017"

var client *mongo.Client

type Config struct {
	Models data.Models
}

func main() {
	mongoClient, err := connectToMongo()
	if err != nil {
		log.Panic(err)
	}

	client = mongoClient

	app := Config{
		Models: data.New(client),
	}

	log.Println("Connected to MongoDB!", app)

	select {}
}

func connectToMongo() (*mongo.Client, error) {
	clientOptions := options.Client().ApplyURI(mongoURL)

	c, err := mongo.Connect(clientOptions)
	if err != nil {
		return nil, err
	}

	return c, nil
}