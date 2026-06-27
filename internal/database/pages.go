package database

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type PageType int8

const (
	pageSize     = 4096
	halfPageSize = pageSize / 2
	maxKeySize   = 256
	maxValueSize = 512
	pageIDSize   = 4
	slotIDSize   = 2

	MetadataPage PageType = 0
	DataPage     PageType = 1
	// B+tree pages
	InternalPage PageType = 2
	LeafPage     PageType = 3
)

// Write new page adding default metadata
func (pages *Pages) writeNewPage(pageID uint32, pageType PageType) error {
	bytes := make([]byte, pageSize)

	pageMetadata := pageMetadata{
		pageType:       pageType,
		pageID:         pageID,
		numEntries:     defaultNumEntries,
		freeSpaceStart: pageMetadataSize,
		freeSpaceEnd:   pageSize,
	}

	buf := createMetadataBuffer(pageMetadata)
	copy(bytes, buf)

	if err := pages.write(bytes, pageID, pageType); err != nil {
		return fmt.Errorf("failed to write page: %w", err)
	}

	return nil
}

// Used to create metedata page for the database. Sets metadata
func (pages *Pages) createMetadataPage() error {
	// Metadata page is always page 1
	const metadataPageID uint32 = 1

	if err := pages.writeNewPage(metadataPageID, MetadataPage); err != nil {
		return err
	}

	metadata := metadata{
		rootPageID:    0,
		totalNumPages: 1, // metadata page exists now
		nextFreePage:  0,
	}

	return pages.updateMetadata(metadata)
}

// Creates all pages
func (pages *Pages) Create(pageType PageType) (uint32, error) {
	var pageID uint32
	if pageType == MetadataPage {
		return 0, errors.New("failed to create page: invalid funcation for creating metadata page")
	}
	metadata, err := pages.ReadMetadata()
	if err != nil {
		return 0, fmt.Errorf("failed to read metadata to create new page: %w", err)
	}
	if metadata.nextFreePage == 0 {
		metadata.totalNumPages++
		pageID = metadata.totalNumPages
	} else {
		pageID = metadata.nextFreePage
		pageIDBytes, err := pages.ReadBytes(pageIDSize, pageStart, metadata.nextFreePage)
		if err != nil {
			return 0, fmt.Errorf("failed to update next free page id")
		}
		// Update next page stack
		nextPageID := binary.BigEndian.Uint32(pageIDBytes)
		metadata.nextFreePage = nextPageID
	}
	if pageType == DataPage {
		metadata.lastDataPage = pageID
	}

	if err := pages.writeNewPage(pageID, pageType); err != nil {
		return 0, err
	}

	if err := pages.updateMetadata(metadata); err != nil {
		return 0, err
	}
	return pageID, nil
}

// Creates the first data page when need and no more
func (pages *Pages) createFirstDataPage() error {
	info, err := pages.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to read database size: %w", err)
	}
	fileSize := info.Size()
	// Checks if only metadata page
	if fileSize == pageSize {
		_, err := pages.Create(DataPage)
		if err != nil {
			return fmt.Errorf("failed to create new data page: %w", err)
		}
	}
	return nil
}

// and convert that into bytes and return it
func getSlotBuffer(newDataOffset uint16, dataLength uint16, flag uint16) []byte {
	buf := make([]byte, slotSize)
	binary.BigEndian.PutUint16(buf[slotOffsetIndex:slotOffsetIndex+slotOffsetSize], newDataOffset)
	binary.BigEndian.PutUint16(buf[slotLengthIndex:slotLengthIndex+slotLengthSize], dataLength)
	binary.BigEndian.PutUint16(buf[slotFlagIndex:slotFlagIndex+slotFlagSize], flag)
	return buf
}

// Writes the len and value of key and the value into a buffer
func getDataBuffer(key string, value string, dataSize int) []byte {
	buf := make([]byte, dataSize)
	var keyLength uint16 = uint16(len(key))
	var valueLength uint16 = uint16(len(value))

	binary.BigEndian.PutUint16(buf[0:2], keyLength)
	binary.BigEndian.PutUint16(buf[2:4], valueLength)

	valueStart := 4 + len(key)
	copy(buf[4:], []byte(key))
	copy(buf[valueStart:], []byte(value))
	return buf
}

// Get next data page to write to and returns new data page info
func (pages *Pages) ensureWritablePage() (pageMetadata, uint32, error) {
	var pageMetadata pageMetadata
	// create a new data page
	pageID, err := pages.Create(DataPage)
	if err != nil {
		return pageMetadata, pageID, fmt.Errorf("failed to add to page %w", err)
	}
	// Compute new page values
	pageMetadata, err = pages.readPageMetadata(pageID)
	if err != nil {
		return pageMetadata, pageID, fmt.Errorf("failed to get metadata: %w", err)
	}
	return pageMetadata, pageID, nil
}
