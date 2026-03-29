package config

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

// MongoClient wraps the mongo client and exposes the target database
type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// NewMongoClient creates and verifies a MongoDB connection
func NewMongoClient(cfg *AppConfig, logger *zap.Logger) (*MongoClient, error) {
	timeout := time.Duration(cfg.MongoTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetConnectTimeout(timeout).
		SetServerSelectionTimeout(timeout)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	logger.Info("MongoDB connected successfully", zap.String("db", cfg.MongoDBName))

	return &MongoClient{
		Client:   client,
		Database: client.Database(cfg.MongoDBName),
	}, nil
}

// Disconnect gracefully closes the MongoDB connection
func (m *MongoClient) Disconnect(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}

// Collection returns a handle to the given collection
func (m *MongoClient) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
