package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitMongoDB(uri, dbName string) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Printf("WARNING: Unable to connect to MongoDB: %v", err)
		return nil
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Printf("WARNING: Unable to ping MongoDB: %v", err)
		return nil
	}

	fmt.Println("MongoDB connected successfully")
	return client.Database(dbName)
}
