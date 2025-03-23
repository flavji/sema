package updateRepoMessage

import (
	"sema/models/delta"

)

// SyncMessage represents the received sync JSON structure.
type UpdateRepoMessage struct {
	Type  string    `json:"type"`
	ReportID string `json:"reportid"`
	Section string  `json:"section"`
	Contents map[string]delta.Delta `json:"contents"`
}

