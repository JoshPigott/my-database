package database

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	metadataPageID      uint32 = 1 << 0
	rootPageID          uint32 = 1 << 1
	undefinedRootPageID uint32 = 0
	noFreePages         uint32 = 0
	onlyMetadataPage    uint32 = 1

	metadataSize       = 12
	rootPageIDIndex    = 0
	totalNumPagesIndex = 4
	freePageStartIndex = 8
)

type metadata struct {
	rootPageID    uint32
	totalNumPages uint32
	freePageStart uint32
}

// Set set value of metadata page
func newDefaultMetadata() metadata {
	return metadata{
		rootPageID:    undefinedRootPageID,
		totalNumPages: onlyMetadataPage,
		freePageStart: noFreePages,
	}
}

// Creates a metadata page if none
func (DB *DB) ensureMetadataPage() error {
	fileInfo, err := os.Stat(FileName)
	if err != nil {
		return fmt.Errorf("failed to check file info: %w", err)
	}
	// Checks if the file size is right
	if fileInfo.Size()%pageSize != 0 {
		return fmt.Errorf("failed to have vaild pages: %w", err)
	}
	if fileInfo.Size() == 0 {
		DB.createMetadataPage()
		return nil
	}
	return nil
}

// Formats bytes into a metadata struc
func formatMetadata(metadataBytes []byte) metadata {
	return metadata{
		rootPageID:    binary.BigEndian.Uint32(metadataBytes[rootPageIDIndex:totalNumPagesIndex]),
		totalNumPages: binary.BigEndian.Uint32(metadataBytes[totalNumPagesIndex:freePageStartIndex]),
		freePageStart: binary.BigEndian.Uint32(metadataBytes[freePageStartIndex:metadataSize]),
	}
}

func (DB *DB) updateMetadata(metadata metadata) error {
	buf := createMetadataBuffer(metadata)
	if err := DB.WriteBytes(buf, pageMetadataSize, metadataPageID); err != nil {
		return err
	}
	return nil
}

// Create the metadata buffer
func createMetadataBuffer(metadata metadata) []byte {
	buf := make([]byte, metadataSize)
	binary.BigEndian.PutUint32(buf[rootPageIDIndex:totalNumPagesIndex], metadata.rootPageID)
	binary.BigEndian.PutUint32(buf[totalNumPagesIndex:freePageStartIndex], metadata.totalNumPages)
	binary.BigEndian.PutUint32(buf[freePageStartIndex:metadataSize], metadata.freePageStart)
	return buf
}

// Return the number of pages that should be read
func (DB *DB) getNumOfPagesToRead() (int, error) {
	metadata, err := DB.readMetadata()
	if err != nil {
		return 0, fmt.Errorf("failed to read number of pages: %w", err)
	}
	if metadata.freePageStart != 0 {
		return int(metadata.freePageStart), nil
	}
	return int(metadata.totalNumPages), nil
}
