package indexer

import (
	"context"
	"errors"

	"agent-cluster/internal/storage"
)

type ResultsResponse struct {
	OK      bool                  `json:"ok"`
	Results []storage.SearchResult `json:"results"`
}

func QueryByName(ctx context.Context, dbPath string, name string, limit int) (ResultsResponse, error) {
	if name == "" {
		return ResultsResponse{}, errors.New("missing name")
	}
	db, err := storage.Open(ctx, dbPath)
	if err != nil {
		return ResultsResponse{}, err
	}
	defer db.Close()

	results, err := storage.QuerySymbolsByName(ctx, db, name, limit)
	if err != nil {
		return ResultsResponse{}, err
	}
	if results == nil {
		results = []storage.SearchResult{}
	}
	return ResultsResponse{OK: true, Results: results}, nil
}

func SearchText(ctx context.Context, dbPath string, query string, limit int) (ResultsResponse, error) {
	if query == "" {
		return ResultsResponse{}, errors.New("missing query")
	}
	db, err := storage.Open(ctx, dbPath)
	if err != nil {
		return ResultsResponse{}, err
	}
	defer db.Close()

	results, err := storage.SearchSymbolsFTS(ctx, db, query, limit)
	if err != nil {
		return ResultsResponse{}, err
	}
	if results == nil {
		results = []storage.SearchResult{}
	}
	return ResultsResponse{OK: true, Results: results}, nil
}
