package database

import (
	"encoding/binary"
	"errors"
	"fmt"
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

// Reads metadata page and return the database metadata in a metadata struct
func (pages *Pages) readDBMetadata() (metadata, error) {
	dbMetadataStart := pageMetadataSize
	metadataBytes, err := pages.ReadBytes(metadataSize, dbMetadataStart, metadataPageID)
	if err != nil {
		return metadata{}, fmt.Errorf("failed to read metadata page: %w", err)
	}
	metadata := formatDBMetadata(metadataBytes)
	return metadata, nil
}

// Creates a metadata page if none
func (pages *Pages) ensureDBMetadataPage(file *os.File) error {
	fileInfo, err := os.Stat(file.Name())
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}
	// Checks if the file size is right (file not corrupted)
	if fileInfo.Size()%pageSize != 0 {
		return errors.New("corrupted file invalid size")
	}
	if fileInfo.Size() == 0 {
		if err := pages.createMetadataPage(); err != nil {
			return fmt.Errorf("failed to create metadata page: %w", err)
		}
		return nil
	}
	return nil
}

// Formats bytes into a metadata struct
func formatDBMetadata(metadataBytes []byte) metadata {
	return metadata{
		rootPageID:    binary.BigEndian.Uint32(metadataBytes[rootPageIDIndex:totalNumPagesIndex]),
		totalNumPages: binary.BigEndian.Uint32(metadataBytes[totalNumPagesIndex:nextFreePageIndex]),
		nextFreePage:  binary.BigEndian.Uint32(metadataBytes[nextFreePageIndex:lastDataPageIndex]),
		lastDataPage:  binary.BigEndian.Uint32(metadataBytes[lastDataPageIndex:metadataSize]),
	}
}

func (pages *Pages) updateDBMetadata(metadata metadata) error {
	buf := createDBMetadataBuffer(metadata)
	if err := pages.writeBytes(buf, pageMetadataSize, metadataPageID); err != nil {
		return err
	}
	return nil
}

// create the metadata buffer
func createDBMetadataBuffer(metadata metadata) []byte {
	buf := make([]byte, metadataSize)
	binary.BigEndian.PutUint32(buf[rootPageIDIndex:totalNumPagesIndex], metadata.rootPageID)
	binary.BigEndian.PutUint32(buf[totalNumPagesIndex:nextFreePageIndex], metadata.totalNumPages)
	binary.BigEndian.PutUint32(buf[nextFreePageIndex:lastDataPageIndex], metadata.nextFreePage)
	binary.BigEndian.PutUint32(buf[lastDataPageIndex:metadataSize], metadata.lastDataPage)
	return buf
}

// Return the number of pages that should be read
func (pages *Pages) getNumOfPagesToRead() (int, error) {
	metadata, err := pages.readDBMetadata()
	if err != nil {
		return 0, fmt.Errorf("failed to read database metadata: %w", err)
	}
	return int(metadata.totalNumPages), nil
}

// Creates or updates the link between the metadata page and the b+tree root
func (root *node) rootPageLink() error {
	buf := make([]byte, rootpageIDSize)
	binary.BigEndian.PutUint32(buf, root.pageID)
	if err := root.pages.writeBytes(buf, pageMetadataSize, metadataPageID); err != nil {
		return fmt.Errorf("failed to write b+root database metadata: %w", err)
	}
	return nil
}

func (pages *Pages) getRootPage() (uint32, bool, error) {
	offset := pageMetadataSize + rootPageIDIndex
	bytes, err := pages.ReadBytes(rootpageIDSize, offset, metadataPageID)
	if err != nil {
		return 0, false, fmt.Errorf("failed to read root page id: %w", err)
	}
	rootID := binary.BigEndian.Uint32(bytes[0:pageIDSize])
	isRootPage := rootID != undefinedRootPageID
	return rootID, isRootPage, err
}
