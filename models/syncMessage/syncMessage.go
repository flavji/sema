package syncMessage

import (
	"sema/models/delta"

)


// SyncMessage represents the received sync JSON structure.
type SyncMessage struct {
	Type  string    `json:"type"`
	ReportID string `json:"reportid"`
	Section string  `json:"section"`
	Contents map[string]delta.Delta `json:"contents"`
}
