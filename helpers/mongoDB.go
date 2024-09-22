package helpers

import (
	"context"
	"log/slog"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InitializeMongo(logger *slog.Logger) *mongo.Database {
	mongoURI := os.Getenv("MONGO_URI")
	dbName := os.Getenv("MONGO_DB_NAME")

	clientOptions := options.Client().ApplyURI(mongoURI)

	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		logger.Error("Failed to Connect to MongoDB", "ERROR", err)
		return nil
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		logger.Error("Failed to Ping MongoDB", "ERROR", err)
		return nil
	}

	logger.Info("Connected to MongoDB Successfully", "Credentials", map[string]string{
		"MONGO_URI":     mongoURI,
		"MONGO_DB_NAME": dbName,
	})

	collection := client.Database(dbName)

	return collection
}
