package data

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var client *mongo.Client

func New(mongoClient *mongo.Client) Models {
	client = mongoClient

	return Models{
		LogEntry: LogEntry{},
	}
}

type Models struct {
	LogEntry LogEntry
}

type LogEntry struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name"`
	Data      string        `bson:"data" json:"data"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}

func (l *LogEntry) Insert(entry LogEntry) error {
	collection := client.Database("logs").Collection("logs")

	_, err := collection.InsertOne(context.TODO(), LogEntry{
		Name:      entry.Name,
		Data:      entry.Data,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err != nil {
		log.Println("Error inserting:", err)
		return err
	}

	return nil
}

func (l *LogEntry) All() ([]*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println("Find error:", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var logs []*LogEntry

	for cursor.Next(ctx) {
		var item LogEntry

		err := cursor.Decode(&item)
		if err != nil {
			log.Println("Decode error:", err)
			return nil, err
		}

		logs = append(logs, &item)
	}

	return logs, nil
}

func (l *LogEntry) GetOne(id string) (*LogEntry, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	docID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var entry LogEntry

	err = client.Database("logs").Collection("logs").
		FindOne(ctx, bson.M{"_id": docID}).
		Decode(&entry)

	if err != nil {
		return nil, err
	}

	return &entry, nil
}

func (l *LogEntry) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	collection := client.Database("logs").Collection("logs")

	if err := collection.Drop(ctx); err != nil {
		return fmt.Errorf("drop collection failed: %w", err)
	}

	return nil
}

func (l *LogEntry) Update(id string) (*mongo.UpdateResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	docID, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	result, err := client.Database("logs").Collection("logs").UpdateOne(
		ctx,
		bson.M{"_id": docID},
		bson.M{
			"$set": bson.M{
				"name":       l.Name,
				"data":       l.Data,
				"updated_at": time.Now(),
			},
		},
	)

	if err != nil {
		return nil, err
	}

	return result, nil
}