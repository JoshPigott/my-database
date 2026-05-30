package database

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type PageType int8

const (
	FileName       = "data/bubbly.db"
	pageSize       = 4096
	pageIDsize int = 4

	MetadataPage PageType = 0
	RootPage     PageType = 1
	RoutingPage  PageType = 2
	DataPage     PageType = 3
)

// Write new page adding default metadata
func (DB *DB) writeNewPage(pageID uint32, pageType PageType) error {
	bytes := make([]byte, pageSize)

	pageMetadata := pageMetadata{
		pageType:       pageType,
		pageID:         pageID,
		numslots:       defaultNumSlots,
		freeSpaceStart: pageMetadataSize,
		freeSpaceEnd:   pageSize,
	}

	buf := CreatePageMetadataBuffer(pageMetadata)
	copy(bytes, buf)

	if err := DB.writePage(bytes, pageType); err != nil {
		return fmt.Errorf("failed to write page: %w", err)
	}

	return nil
}

// Used to create metedata page for the database. Sets metadata
func (DB *DB) createMetadataPage() error {
	// Metadata page is always page 1
	const metadataPageID uint32 = 1

	if err := DB.writeNewPage(metadataPageID, MetadataPage); err != nil {
		return err
	}

	metadata := metadata{
		rootPageID:    0,
		totalNumPages: 1, // metadata page exists now
		freePageStart: 0,
	}

	return DB.updateMetadata(metadata)
}

// Creates all pages
func (DB *DB) CreatePage(pageType PageType) error {
	if pageType == MetadataPage {
		return errors.New("failed to create page: invalid funcation for creating metadata page")
	}
	metadata, err := DB.readMetadata()
	if err != nil {
		return err
	}

	metadata.totalNumPages++
	pageID := metadata.totalNumPages
	if err := DB.writeNewPage(pageID, pageType); err != nil {
		return err
	}

	if err := DB.updateMetadata(metadata); err != nil {
		return err
	}
	return nil
}

// Creates the first data page when need and no more
func (DB *DB) createFirstDataPage() error {
	info, err := DB.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to read database size: %w", err)
	}
	fileSize := info.Size()
	// Checks if only metadata page
	if fileSize == pageSize {
		if err := DB.CreatePage(DataPage); err != nil {
			return fmt.Errorf("failed to create new data page: %w", err)
		}
	}
	return nil
}

// Finds the data offset
func getOffsets(freeSpace int, dataSize int, numslots uint16) (int, int) {
	totalSlotSize := int(numslots) * slotSize
	slotOffset := totalSlotSize + pageMetadataSize

	dataStart := slotOffset + freeSpace
	newDataOffset := dataStart - dataSize
	return slotOffset, newDataOffset
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
func (DB *DB) ensureWritablePage(metadata metadata, pageID uint32) (pageMetadata, uint32, error) {
	var pageMetadata pageMetadata
	if metadata.freePageStart != 0 {
		// uses a free page
		pageID = metadata.freePageStart
		// I will still need to update some stuff here
		// But I will do it when I get to free pages
	} else {
		// create a new page
		err := DB.CreatePage(DataPage)
		if err != nil {
			return pageMetadata, pageID, fmt.Errorf("failed to add to page %w", err)
		}
		// Compute new page values
		pageID++
		pageMetadata, err = DB.readPageMetadata(pageID)
		if err != nil {
			return pageMetadata, pageID, fmt.Errorf("failed to get metadata: %w", err)
		}
		return pageMetadata, pageID, nil
	}
	return pageMetadata, pageID, nil
}
