package indexer

import (
	"context"
	"errors"

	"agent-cluster/internal/storage"
)

type RefsResponse struct {
	OK      bool             `json:"ok"`
	Results []storage.RefRow `json:"results"`
}

func RefsBySymbolID(ctx context.Context, dbPath string, symbolID int64, limit int) (RefsResponse, error) {
	if symbolID <= 0 {
		return RefsResponse{}, errors.New("missing --id")
	}
	db, err := storage.Open(ctx, dbPath)
	if err != nil {
		return RefsResponse{}, err
	}
	defer db.Close()

	results, err := storage.ListRefsBySymbolID(ctx, db, symbolID, limit)
	if err != nil {
		return RefsResponse{}, err
	}
	if results == nil {
		results = []storage.RefRow{}
	}
	return RefsResponse{OK: true, Results: results}, nil
}

