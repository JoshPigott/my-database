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
	numEntriesIndex     = 5
	freeSpaceStartIndex = 7
	freeSpaceEndIndex   = 9

	defaultNumEntries uint16 = 0
)

type pageMetadata struct {
	pageType       PageType
	pageID         uint32
	numEntries     uint16
	freeSpaceStart uint16
	freeSpaceEnd   uint16
}

// Rewrite page metadata
func (Pages *Pages) updatePageMetadata(pageMetadata pageMetadata) error {
	buf := createMetadataBuffer(pageMetadata)
	err := Pages.WriteBytes(buf, 0, pageMetadata.pageID)
	if err != nil {
		return fmt.Errorf("failed to update page metadata: %w", err)
	}
	return nil
}

// Formats bytes into a metadata struc
func formatPageMetadata(metadataBytes []byte) pageMetadata {
	return pageMetadata{
		pageType:       PageType(int8(metadataBytes[pageTypeIndex])),
		pageID:         binary.BigEndian.Uint32(metadataBytes[pageIDIndex:numEntriesIndex]),
		numEntries:     binary.BigEndian.Uint16(metadataBytes[numEntriesIndex:freeSpaceStartIndex]),
		freeSpaceStart: binary.BigEndian.Uint16(metadataBytes[freeSpaceStartIndex:freeSpaceEndIndex]),
		freeSpaceEnd:   binary.BigEndian.Uint16(metadataBytes[freeSpaceEndIndex:pageMetadataSize]),
	}
}

// Create the metadata buffer
func createMetadataBuffer(pageMetadata pageMetadata) []byte {
	buf := make([]byte, pageMetadataSize)
	buf[pageTypeIndex] = byte(pageMetadata.pageType)
	binary.BigEndian.PutUint32(buf[pageIDIndex:numEntriesIndex], pageMetadata.pageID)
	binary.BigEndian.PutUint16(buf[numEntriesIndex:freeSpaceStartIndex], pageMetadata.numEntries)
	binary.BigEndian.PutUint16(buf[freeSpaceStartIndex:freeSpaceEndIndex], pageMetadata.freeSpaceStart)
	binary.BigEndian.PutUint16(buf[freeSpaceEndIndex:pageMetadataSize], pageMetadata.freeSpaceEnd)
	return buf
}
