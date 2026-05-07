package database

import (
	"encoding/json"
	"fmt"
	"os"
)

const metadataFileName string = "data/metadata.json"
const temporaryFile string = "data/metadata-temp.json"

type Metadata struct {
	// File                          *os.File `json:"-"` // Excluded from JSON file
	OperationsSinceLastCompaction int `json:"operation_since_last_compaction"`
}

// Read meta from metadata file and meatadata struc
func getMetadata() (*Metadata, error) {
	// Read file
	content, err := os.ReadFile(metadataFileName)
	if err != nil {
		return &Metadata{}, fmt.Errorf("failed to read metadata file: %w", err)
	}
	// Phrase json content
	metadata := Metadata{}
	err = json.Unmarshal(content, &metadata)
	if err != nil {
		return &Metadata{}, fmt.Errorf("failed to phrase metadata file: %w", err)
	}
	return &metadata, err
}

// Update metadata file using a temp file
func (m *Metadata) Update() error {
	content, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to matshal metadata: %w", err)
	}
	tempFile, err := os.Create(temporaryFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary metadata file: %w", err)
	}
	if _, err = tempFile.Write(content); err != nil {
		return fmt.Errorf("failed to write metadata to temporary file: %w", err)
	}
	if err = tempFile.Sync(); err != nil {
		return fmt.Errorf("failed to write metadata to temporary file: %w", err)
	}
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary metadata file: %w", err)
	}
	if err = os.Rename(temporaryFile, metadataFileName); err != nil {
		return fmt.Errorf("failed to switch metafile with temporary file: %w", err)
	}
	return nil
}
