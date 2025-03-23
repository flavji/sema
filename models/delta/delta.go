package delta

import (
	"encoding/json"
)

// Delta represents the structure of a delta object that includes editorId and operations.
type Delta struct {
    Type  string    `json:"type"` // Type of the message, e.g., "delta"
    Delta DeltaData `json:"delta"` // The delta (editorId + operations)
}

// DeltaData represents the data for the delta, including the editorId and the actual delta operations.
type DeltaData struct {
    EditorId string   `json:"editorId"` // The ID of the editor that the delta is related to
    Delta    DeltaOps `json:"delta"`    // The delta operations to be applied
}

// DeltaOps defines the operations that can be applied in the delta.
type DeltaOps struct {
    Ops []DeltaOp `json:"ops"` // A slice of operations like insert, retain, delete
}

// DeltaOp represents a single operation in the delta
type DeltaOp struct {
    Retain int    `json:"retain,omitempty"` // Retain a number of characters
		Insert json.RawMessage `json:"insert,omitempty"`
    Delete int    `json:"delete,omitempty"` // Delete a number of characters
		Attributes *Attributes     `json:"attributes,omitempty"` // Text formatting (bold, italic, etc.)
}


// Attributes represents the formatting options available in Quill.
type Attributes struct {
    Bold   *bool  `json:"bold,omitempty"`   // Bold text
    Italic *bool  `json:"italic,omitempty"` // Italic text
		List  *string `json:"list"` 	// Lists
    Underline *bool  `json:"underline,omitempty"` // Underline text
    Link   *string `json:"link,omitempty"`   // Hyperlink
    // Color  *string `json:"color,omitempty"`  // Text color
    // Font   *string `json:"font,omitempty"`   // Font type
    // Size   *string `json:"size,omitempty"`   // Font size
}

/* ImageEmbed represents an embedded image in the Quill delta. */
type ImageEmbed struct {
    Image string `json:"image"` // Image URL
}
