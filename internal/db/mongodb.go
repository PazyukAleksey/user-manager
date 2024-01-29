package db

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mongoInit() string {
	err := godotenv.Load(".env")
	if err != nil {
		log.Panic("Error loading .env file")
	}
	return fmt.Sprintf("mongodb+srv://%s:%s@%s", os.Getenv("db_account"), os.Getenv("db_password"), os.Getenv("db_url"))
}

func ConnectToMongo() (*mongo.Client, error) {
	mongoUri := mongoInit()
	fmt.Println(mongoUri)
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoUri))
	if err != nil {
		return nil, fmt.Errorf("mongo connect error: %w", err)
	}
	return client, nil
}
