package database

import (
	"encoding/binary"
	"fmt"
)

const (
	// Indexs and sizes
	pageMetadataSize    = 11
	pageTypeIndex       = 0
	pageIDIndex         = 1
	numSlotsIndex       = 5
	freeSpaceStartIndex = 7
	freeSpaceEndIndex   = 9

	defaultNumSlots uint16 = 0
)

type pageMetadata struct {
	pageType       PageType
	pageID         uint32
	numSlots       uint16
	freeSpaceStart uint16
	freeSpaceEnd   uint16
}

// Rewrite page metadata
func (DB *DB) updatePageMetadata(oldPageMetadata pageMetadata, dataSize int) error {
	newPageMetadata := pageMetadata{
		pageType:       oldPageMetadata.pageType,
		pageID:         oldPageMetadata.pageID,
		numSlots:       oldPageMetadata.numSlots + 1,
		freeSpaceStart: oldPageMetadata.freeSpaceStart + uint16(slotSize),
		freeSpaceEnd:   oldPageMetadata.freeSpaceEnd - uint16(dataSize),
	}
	buf := CreatePageMetadataBuffer(newPageMetadata)
	err := DB.WriteBytes(buf, 0, oldPageMetadata.pageID)
	if err != nil {
		return fmt.Errorf("failed to update page metadata: %w", err)
	}
	return err
}

// Formats bytes into a metadata struc
func formatPageMetadata(metadataBytes []byte) pageMetadata {
	return pageMetadata{
		pageType:       PageType(int8(metadataBytes[pageTypeIndex])),
		pageID:         binary.BigEndian.Uint32(metadataBytes[pageIDIndex:numSlotsIndex]),
		numSlots:       binary.BigEndian.Uint16(metadataBytes[numSlotsIndex:freeSpaceStartIndex]),
		freeSpaceStart: binary.BigEndian.Uint16(metadataBytes[freeSpaceStartIndex:freeSpaceEndIndex]),
		freeSpaceEnd:   binary.BigEndian.Uint16(metadataBytes[freeSpaceEndIndex:pageMetadataSize]),
	}
}

// Create the metadata buffer
func CreatePageMetadataBuffer(pageMetadata pageMetadata) []byte {
	buf := make([]byte, pageMetadataSize)
	buf[pageTypeIndex] = byte(pageMetadata.pageType)
	binary.BigEndian.PutUint32(buf[pageIDIndex:numSlotsIndex], pageMetadata.pageID)
	binary.BigEndian.PutUint16(buf[numSlotsIndex:freeSpaceStartIndex], pageMetadata.numSlots)
	binary.BigEndian.PutUint16(buf[freeSpaceStartIndex:freeSpaceEndIndex], pageMetadata.freeSpaceStart)
	binary.BigEndian.PutUint16(buf[freeSpaceEndIndex:pageMetadataSize], pageMetadata.freeSpaceEnd)
	return buf
}
