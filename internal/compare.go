package internal

import (
	"fmt"
	"github.com/fadyat/mongo-cmp/cmd/flags"
	"github.com/fadyat/mongo-cmp/internal/api"
	"github.com/fadyat/mongo-cmp/internal/progress"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"time"
)

var noStats = stats{}

var (
	Skipped                    = "skipped"
	FailedToGetCollectionStats = "failed to get collection stats"
	FailedToCountDocuments     = "failed to count documents"
	NotFound                   = "not found"
	Succeeded                  = "succeeded"
)

type collectionStats struct {
	Name            string
	DocumentsNumber int64
	Stats           bson.M
	Status          string
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

type mongoService struct {
	name string
	api  api.MongoApi
}

func newMongoService(name string, c api.MongoApi) *mongoService {
	return &mongoService{
		name: name,
		api:  c,
	}
}

func (m *mongoService) ListDatabaseNames(f *flags.CompareFlags) ([]string, error) {
	if f.Database == flags.DefaultDatabase {
		return m.api.ListDatabaseNames()
	}

	return []string{f.Database}, nil
}

func (m *mongoService) CollectStats(f *flags.CompareFlags) (stats, error) {
	databases, err := m.ListDatabaseNames(f)
	if err != nil {
		return noStats, err
	}

	var result = newStats()
	result.Databases = databases

	for _, dbName := range databases {
		collections, err := m.api.ListCollectionNames(dbName)
		if err != nil {
			return noStats, err
		}

		header := fmt.Sprintf("%s(%s): %d collections", dbName, m.name, len(collections))
		pb := progress.NewProgress(header)
		pb.Start(len(collections))

		for _, collectionName := range collections {
			pb.Increment()

			if result.Collections[dbName] == nil {
				result.Collections[dbName] = make(map[string]collectionStats)
			}

			if !f.ShowDetails {
				result.Collections[dbName][collectionName] = collectionStats{
					Name:   collectionName,
					Status: Skipped,
				}
				continue
			}

			s := m.aggregateCollectionStats(dbName, collectionName)
			result.Collections[dbName][collectionName] = s
		}

		pb.Finish()
	}

	return result, nil
}

func (m *mongoService) aggregateCollectionStats(dbName, collectionName string) collectionStats {
	meta := map[string]string{
		"database":   dbName,
		"collection": collectionName,
	}

	statistics, err := m.api.CollectionStats(dbName, collectionName)
	if err != nil {
		slog.Error("failed to get collection stats", "meta", meta, "error", err)
		return collectionStats{
			Name:   collectionName,
			Status: FailedToGetCollectionStats,
		}
	}

	count, err := m.api.CountDocuments(dbName, collectionName)
	if err != nil {
		slog.Error("failed to count documents", "meta", meta, "error", err)
		return collectionStats{
			Name:   collectionName,
			Stats:  statistics,
			Status: FailedToCountDocuments,
		}
	}

	return collectionStats{
		Name:            collectionName,
		DocumentsNumber: count,
		Stats:           statistics,
		Status:          Succeeded,
	}
}

func collectData(f *flags.CompareFlags) (*clusterStats, error) {
	sourceApi, err := api.NewMongoApi(f.From, f.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the source (from) database: %w", err)
	}

	destinationApi, err := api.NewMongoApi(f.To, f.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the destination (to) database: %w", err)
	}

	var (
		sourceStats, destinationStats stats
		errGroup                      errgroup.Group
	)

	sourceService := newMongoService("from", sourceApi)
	errGroup.Go(func() (err error) {
		sourceStats, err = sourceService.CollectStats(f)
		if err != nil {
			return fmt.Errorf("failed to collect stats from the source database: %w", err)
		}

		return nil
	})

	time.Sleep(200 * time.Millisecond)

	destinationService := newMongoService("to", destinationApi)
	errGroup.Go(func() (err error) {
		destinationStats, err = destinationService.CollectStats(f)
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
