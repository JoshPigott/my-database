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
func (pages *Pages) updatePageMetadata(pm pageMetadata) error {
	buf := createMetadataBuffer(pm)
	err := pages.writeBytes(buf, 0, pm.pageID)
	if err != nil {
		return fmt.Errorf("failed to write page metadata: %w", err)
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

// create the metadata buffer
func createMetadataBuffer(pm pageMetadata) []byte {
	buf := make([]byte, pageMetadataSize)
	buf[pageTypeIndex] = byte(pm.pageType)
	binary.BigEndian.PutUint32(buf[pageIDIndex:numEntriesIndex], pm.pageID)
	binary.BigEndian.PutUint16(buf[numEntriesIndex:freeSpaceStartIndex], pm.numEntries)
	binary.BigEndian.PutUint16(buf[freeSpaceStartIndex:freeSpaceEndIndex], pm.freeSpaceStart)
	binary.BigEndian.PutUint16(buf[freeSpaceEndIndex:pageMetadataSize], pm.freeSpaceEnd)
	return buf
}
