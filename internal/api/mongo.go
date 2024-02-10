package api

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log/slog"
	"time"
)

const (
	// InitTimeout is the timeout for initializing the MongoDB connection
	InitTimeout = 10 * time.Second
)

type MongoApi interface {
	ListDatabaseNames() ([]string, error)

	ListCollectionNames(dbName string) ([]string, error)

	CountDocuments(dbName string, collectionName string) (int64, error)

	CollectionStats(dbName, collectionName string) (bson.M, error)
}

func initMongoConnection(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), InitTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}

	return client, nil
}

type mongoApi struct {
	name             string
	operationTimeout time.Duration
	Client           *mongo.Client
}

func NewMongoApi(name, uri string, operationTimeout time.Duration) (MongoApi, error) {
	client, err := initMongoConnection(uri)
	if err != nil {
		return nil, err
	}

	return &mongoApi{
		name:             name,
		Client:           client,
		operationTimeout: operationTimeout,
	}, nil
}

func (m *mongoApi) ListDatabaseNames() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.operationTimeout)
	defer cancel()

	slog.Debug("listing databases", "where", m.name)
	return m.Client.ListDatabaseNames(ctx, bson.D{})
}

func (m *mongoApi) ListCollectionNames(dbName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.operationTimeout)
	defer cancel()

	slog.Debug("listing collections", "where", m.name, "database", dbName)
	return m.Client.Database(dbName).ListCollectionNames(ctx, bson.D{}, options.ListCollections())
}

func (m *mongoApi) CountDocuments(dbName, collectionName string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.operationTimeout)
	defer cancel()

	// ignoring system.sessions collection, we are not interested in it
	if collectionName == "system.sessions" {
		return 0, nil
	}

	// https://www.mongodb.com/docs/drivers/go/current/fundamentals/crud/read-operations/count/#accurate-count
	//
	// Avoiding the sequential scan of the entire collection.
	opts := options.Count().SetHint("_id_")

	slog.Debug("counting documents", "where", m.name, "database", dbName, "collection", collectionName)
	count, err := m.Client.Database(dbName).
		Collection(collectionName).
		CountDocuments(ctx, bson.D{}, opts)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (m *mongoApi) CollectionStats(dbName, collectionName string) (bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), m.operationTimeout)
	defer cancel()

	slog.Debug("getting collection stats", "where", m.name, "database", dbName, "collection", collectionName)
	singleResult := m.Client.Database(dbName).
		RunCommand(ctx, bson.M{"collStats": collectionName})

	if singleResult.Err() != nil {
		return nil, singleResult.Err()
	}

	var result bson.M
	if err := singleResult.Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
