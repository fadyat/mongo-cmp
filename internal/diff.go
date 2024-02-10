package internal

import (
	"encoding/json"
	"fmt"
	"github.com/jedib0t/go-pretty/v6/table"
	"go.mongodb.org/mongo-driver/bson"
	"os"
)

type relevantStats struct {
	CollectionSize any `json:"collection_size"`
	StorageSize    any `json:"storage_size"`
	TotalIndexSize any `json:"total_index_size"`
	TotalSize      any `json:"total_size"`
	DocumentNumber any `json:"document_number"`
}

func getRelevantStats(cstats bson.M, docNumber int64) *relevantStats {
	return &relevantStats{
		CollectionSize: cstats["size"],
		StorageSize:    cstats["storageSize"],
		TotalIndexSize: cstats["totalIndexSize"],
		TotalSize:      cstats["totalSize"],
		DocumentNumber: docNumber,
	}
}

func (s *relevantStats) json() string {
	b, _ := json.MarshalIndent(s, "", "  ")
	return string(b)
}

func boolToSymbol(b bool) string {
	if b {
		return "+"
	}

	return "-"
}

func mergeSlices(s1, s2 []string) []string {
	m := make(map[string]bool)
	for _, item := range s1 {
		m[item] = true
	}

	for _, item := range s2 {
		m[item] = true
	}

	var result = make([]string, 0, len(m))
	for item := range m {
		result = append(result, item)
	}

	return result
}

func getMapKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

func showDatabaseDiff(dbName string, s *clusterStats) {
	source, ok := s.Source.Collections[dbName]
	if !ok {
		source = map[string]collectionStats{}
	}

	destination, ok := s.Destination.Collections[dbName]
	if !ok {
		destination = map[string]collectionStats{}
	}

	mergedCollections := mergeSlices(getMapKeys(source), getMapKeys(destination))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleLight)
	t.AppendHeader(table.Row{"Collection", "Same", "Source", "Destination"})

	for _, collectionName := range mergedCollections {
		sourceStats, have := source[collectionName]
		if !have {
			sourceStats = collectionStats{}
		}

		destinationStats, have := destination[collectionName]
		if !have {
			destinationStats = collectionStats{}
		}

		srelevant := getRelevantStats(sourceStats.Stats, sourceStats.DocumentsNumber)
		drelevant := getRelevantStats(destinationStats.Stats, destinationStats.DocumentsNumber)

		same := boolToSymbol(srelevant == drelevant)
		if sourceStats.Failed || destinationStats.Failed {
			same = "N/A"
		}

		t.AppendRow(table.Row{
			collectionName,
			same,
			getRelevantStats(sourceStats.Stats, sourceStats.DocumentsNumber).json(),
			getRelevantStats(destinationStats.Stats, destinationStats.DocumentsNumber).json(),
		})
		t.AppendSeparator()
	}

	t.Render()
}

func showDiff(s *clusterStats) {
	mergedDatabases := mergeSlices(s.Source.Databases, s.Destination.Databases)
	for _, dbName := range mergedDatabases {
		fmt.Println("Database:", dbName)
		showDatabaseDiff(dbName, s)
		fmt.Println()
	}
}
