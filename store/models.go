package store

import (
	"encoding/json"
	"fmt"

	"github.com/tuongaz/go-saas/store/types"
)

// List represents a collection of records with metadata
type List struct {
	Records []types.Record
	Meta    Metadata
}

// Metadata holds information about a query result
type Metadata struct {
	Total      int `json:"total"`
	Limit      int `json:"limit"`
	Offset     int `json:"offset"`
	TotalPages int `json:"total_pages"`
}

func (l List) Decode(obj any) error {
	jsonData, err := json.Marshal(l.Records)
	if err != nil {
		return fmt.Errorf("encode to json: %w", err)
	}
	if err := json.Unmarshal(jsonData, obj); err != nil {
		return fmt.Errorf("decode json to struct: %w", err)
	}
	return nil
}

// Pagination represents pagination parameters for queries
type Pagination struct {
	Limit  int
	Offset int
}
