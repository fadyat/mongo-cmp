package internal

import (
	"context"
	"fmt"
	"github.com/fadyat/mongo-cmp/cmd/flags"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"time"
)

const (
	// InitTimeout is the timeout for initializing the MongoDB connection
	InitTimeout = 10 * time.Second
)

// OperationTimeout is the timeout for MongoDB operations
// value is overridden by the command line flag "timeout"
var OperationTimeout time.Duration
var noStats = stats{}

type collectionStats struct {
	Name            string
	DocumentsNumber int64
	Stats           bson.M
	Failed          bool
}

type stats struct {
	Databases   []string
	Collections map[string]map[string]collectionStats
}

func newStats() stats {
	return stats{
		Collections: make(map[string]map[string]collectionStats),
		Databases:   make([]string, 0),
	}
}

type clusterStats struct {
	Source      stats
	Destination stats
}

type MongoApi interface {
	ListDatabaseNames() ([]string, error)

	ListCollectionNames(dbName string) ([]string, error)

	CountDocuments(dbName string, collectionName string) (int64, error)

	CollectionStats(dbName, collectionName string) (bson.M, error)

	CollectStats() (stats, error)
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
	name   string
	Client *mongo.Client
}

func newMongoApi(name, uri string) (MongoApi, error) {
	client, err := initMongoConnection(uri)
	if err != nil {
		return nil, err
	}

	return &mongoApi{
		name:   name,
		Client: client,
	}, nil
}

func (m *mongoApi) ListDatabaseNames() ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	slog.Debug("listing databases", "where", m.name)
	return m.Client.ListDatabaseNames(ctx, bson.D{})
}

func (m *mongoApi) ListCollectionNames(dbName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
	defer cancel()

	slog.Debug("listing collections", "where", m.name, "database", dbName)
	return m.Client.Database(dbName).ListCollectionNames(ctx, bson.D{}, options.ListCollections())
}

func (m *mongoApi) CountDocuments(dbName, collectionName string) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
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
	ctx, cancel := context.WithTimeout(context.Background(), OperationTimeout)
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

func (m *mongoApi) CollectStats() (stats, error) {
	databases, e := m.ListDatabaseNames()
	if e != nil {
		return stats{}, e
	}

	var result = newStats()
	result.Databases = databases

	for _, dbName := range databases {
		collections, err := m.ListCollectionNames(dbName)
		if err != nil {
			return noStats, err
		}

		for _, collectionName := range collections {
			s := m.aggregateCollectionStats(dbName, collectionName)

			if result.Collections[dbName] == nil {
				result.Collections[dbName] = make(map[string]collectionStats)
			}

			result.Collections[dbName][collectionName] = s
		}
	}

	return result, nil
}

func (m *mongoApi) aggregateCollectionStats(dbName, collectionName string) collectionStats {
	meta := map[string]string{
		"database":   dbName,
		"collection": collectionName,
	}

	statistics, err := m.CollectionStats(dbName, collectionName)
	if err != nil {
		slog.Error("failed to get collection stats", "meta", meta, "error", err)
		return collectionStats{
			Name:   collectionName,
			Failed: true,
		}
	}

	count, err := m.CountDocuments(dbName, collectionName)
	if err != nil {
		slog.Error("failed to count documents", "meta", meta, "error", err)
		return collectionStats{
			Name:   collectionName,
			Stats:  statistics,
			Failed: true,
		}
	}

	return collectionStats{
		Name:            collectionName,
		DocumentsNumber: count,
		Stats:           statistics,
	}
}

func (m *mongoApi) Close() error {
	return m.Client.Disconnect(context.Background())
}

func collectData(f *flags.CompareFlags) (*clusterStats, error) {
	OperationTimeout = f.Timeout

	sourceApi, err := newMongoApi("from", f.From)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the source database: %w", err)
	}

	destinationApi, err := newMongoApi("to", f.To)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the destination database: %w", err)
	}

	var (
		sourceStats, destinationStats stats
		errGroup                      errgroup.Group
	)

	errGroup.Go(func() (err error) {
		sourceStats, err = sourceApi.CollectStats()
		if err != nil {
			return fmt.Errorf("failed to collect stats from the source database: %w", err)
		}

		return nil
	})

	errGroup.Go(func() (err error) {
		destinationStats, err = destinationApi.CollectStats()
		if err != nil {
			return fmt.Errorf("failed to collect stats from the destination database: %w", err)
		}

		return nil
	})

	if err := errGroup.Wait(); err != nil {
		return nil, err
	}

	return &clusterStats{
		Source:      sourceStats,
		Destination: destinationStats,
	}, nil
}

func Compare(f *flags.CompareFlags) error {
	data, err := collectData(f)
	if err != nil {
		return err
	}

	showDiff(data)
	return nil
}
