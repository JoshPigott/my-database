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
	pageTypeSize = 1

	FreePage     PageType = 0
	MetadataPage PageType = 1
	DataPage     PageType = 2
	// B+tree pages
	InternalPage PageType = 3
	LeafPage     PageType = 4
)

// Creates all pages
func (pages *Pages) create(pageType PageType) (uint32, error) {
	if pageType == MetadataPage {
		return 0, errors.New("failed to create page: invalid funcation for creating metadata page")
	}
	pageID, metadata, err := pages.selectPageId()
	if err != nil {
		return 0, fmt.Errorf("failed to select page id: %w", err)
	}
	if pageType == DataPage {
		metadata.lastDataPage = pageID
	}
	if err := pages.writeNewPage(pageID, pageType); err != nil {
		return 0, err
	}

	if err := pages.updateDBMetadata(metadata); err != nil {
		return 0, err
	}
	return pageID, nil
}

// With database metadata selects the id for the new page
func (pages *Pages) selectPageId() (uint32, metadata, error) {
	var pageID uint32
	metadata, err := pages.readDBMetadata()
	if err != nil {
		return 0, metadata, fmt.Errorf("failed to read metadata: %w", err)
	}
	if metadata.nextFreePage == 0 {
		metadata.totalNumPages++
		pageID = metadata.totalNumPages
	} else {
		pageID = metadata.nextFreePage
		// Update next page stack
		pageIDBytes, err := pages.ReadBytes(pageIDSize, pageTypeIndex, metadata.nextFreePage)
		if err != nil {
			return 0, metadata, fmt.Errorf("failed to get old first free page id: %w", err)
		}
		nextPageID := binary.BigEndian.Uint32(pageIDBytes)
		metadata.nextFreePage = nextPageID
	}
	return pageID, metadata, nil
}

// Write new page adding default metadata
func (pages *Pages) writeNewPage(pageID uint32, pageType PageType) error {
	pageBytes := make([]byte, pageSize)

	pageMetadata := pageMetadata{
		pageType:       pageType,
		pageID:         pageID,
		numEntries:     defaultNumEntries,
		freeSpaceStart: pageMetadataSize,
		freeSpaceEnd:   pageSize,
	}

	buf := createMetadataBuffer(pageMetadata)
	copy(pageBytes, buf)

	info, err := pages.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	if pageType == MetadataPage && info.Size() != 0 {
		return errors.New("failed to add metadata page: metadata already made")
	}
	if err := pages.writeBytes(pageBytes, pageStart, pageID); err != nil {
		return fmt.Errorf("failed to write page bytes: %w", err)
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

	return pages.updateDBMetadata(metadata)
}

// Creates the first data page when need and no more
func (pages *Pages) createFirstDataPage() error {
	info, err := pages.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file stats: %w", err)
	}
	fileSize := info.Size()
	// Checks if databse has been set up before
	if fileSize == 0 {
		_, err := pages.create(DataPage)
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
	keyLength := uint16(len(key))
	valueLength := uint16(len(value))

	binary.BigEndian.PutUint16(buf[0:2], keyLength)
	binary.BigEndian.PutUint16(buf[2:4], valueLength)

	valueStart := 4 + len(key)
	copy(buf[4:], []byte(key))
	copy(buf[valueStart:], []byte(value))
	return buf
}

// Get next data page to write to and returns new data page info
func (pages *Pages) ensureWritablePage() (pageMetadata, uint32, error) {
	// create a new data page
	pageID, err := pages.create(DataPage)
	if err != nil {
		return pageMetadata{}, pageID, fmt.Errorf("failed to create data page %w", err)
	}
	// Compute new page values
	pageMetadata, err := pages.readPageMetadata(pageID)
	if err != nil {
		return pageMetadata, pageID, fmt.Errorf("failed to read database metadata: %w", err)
	}
	return pageMetadata, pageID, nil
}
