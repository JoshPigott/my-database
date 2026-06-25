package database

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	metadataPageID      uint32 = 1 << 0
	undefinedRootPageID uint32 = 0
	noFreePages         uint32 = 0
	onlyMetadataPage    uint32 = 1

	metadataSize       = 16
	rootpageIDSize     = 4
	rootPageIDIndex    = 0
	totalNumPagesIndex = 4
	nextFreePageIndex  = 8
	lastDataPageIndex  = 12
)

type metadata struct {
	rootPageID    uint32
	totalNumPages uint32
	nextFreePage  uint32
	lastDataPage  uint32
}

// Set set value of metadata page
func newDefaultMetadata() metadata {
	return metadata{
		rootPageID:    undefinedRootPageID,
		totalNumPages: onlyMetadataPage,
		nextFreePage:  noFreePages,
	}
}

// Reads metadata page and return the database metadata in a metadata struct
func (Pages *Pages) ReadMetadata() (metadata, error) {
	metadataBytes := make([]byte, metadataSize)
	dbMetadataStart := int64(pageMetadataSize)
	if _, err := Pages.File.Seek(dbMetadataStart, io.SeekStart); err != nil {
		return metadata{}, fmt.Errorf("failed to read metadata page: %w", err)
	}
	if _, err := Pages.File.Read(metadataBytes); err != nil {
		return metadata{}, fmt.Errorf("failed to read metadata page: %w", err)
	}
	metadata := formatMetadata(metadataBytes)
	return metadata, nil
}

// Creates a metadata page if none
func (Pages *Pages) ensureMetadataPage() error {
	fileInfo, err := os.Stat(FileName)
	if err != nil {
		return fmt.Errorf("failed to check file info: %w", err)
	}
	// Checks if the file size is right (file not corrupted)
	if fileInfo.Size()%pageSize != 0 {
		return fmt.Errorf("failed to have vaild pages: %w", err)
	}
	if fileInfo.Size() == 0 {
		Pages.createMetadataPage()
		return nil
	}
	return nil
}

// Formats bytes into a metadata struct
func formatMetadata(metadataBytes []byte) metadata {
	return metadata{
		rootPageID:    binary.BigEndian.Uint32(metadataBytes[rootPageIDIndex:totalNumPagesIndex]),
		totalNumPages: binary.BigEndian.Uint32(metadataBytes[totalNumPagesIndex:nextFreePageIndex]),
		nextFreePage:  binary.BigEndian.Uint32(metadataBytes[nextFreePageIndex:lastDataPageIndex]),
		lastDataPage:  binary.BigEndian.Uint32(metadataBytes[lastDataPageIndex:metadataSize]),
	}
}

func (Pages *Pages) updateMetadata(metadata metadata) error {
	buf := createDatabaseMetadataBuffer(metadata)
	if err := Pages.WriteBytes(buf, pageMetadataSize, metadataPageID); err != nil {
		return err
	}
	return nil
}

// Create the metadata buffer
func createDatabaseMetadataBuffer(metadata metadata) []byte {
	buf := make([]byte, metadataSize)
	binary.BigEndian.PutUint32(buf[rootPageIDIndex:totalNumPagesIndex], metadata.rootPageID)
	binary.BigEndian.PutUint32(buf[totalNumPagesIndex:nextFreePageIndex], metadata.totalNumPages)
	binary.BigEndian.PutUint32(buf[nextFreePageIndex:lastDataPageIndex], metadata.nextFreePage)
	binary.BigEndian.PutUint32(buf[lastDataPageIndex:metadataSize], metadata.lastDataPage)
	return buf
}

// Return the number of pages that should be read
func (Pages *Pages) getNumOfPagesToRead() (int, error) {
	metadata, err := Pages.ReadMetadata()
	if err != nil {
		return 0, fmt.Errorf("failed to read number of pages: %w", err)
	}
	return int(metadata.totalNumPages), nil
}

// Creates or updates the link between the metadata page and the b+tree root
func (Pages *Pages) rootPageLink(newRootID uint32) error {
	buf := make([]byte, rootpageIDSize)
	binary.BigEndian.PutUint32(buf, newRootID)
	if err := Pages.WriteBytes(buf, rootPageIDIndex, metadataPageID); err != nil {
		return fmt.Errorf("failed to create link between metadata page and b+tree root page")
	}
	return nil
}

func (Pages *Pages) getRootPage() (uint32, bool, error) {
	bytes, err := Pages.ReadBytes(rootpageIDSize, rootPageIDIndex, metadataPageID)
	if err != nil {
		return 0, false, fmt.Errorf("failed to read root page id: %w", err)
	}
	rootID := binary.BigEndian.Uint32(bytes[0:pageIDSize])
	isRootPage := rootID != undefinedRootPageID
	return rootID, isRootPage, err
}
