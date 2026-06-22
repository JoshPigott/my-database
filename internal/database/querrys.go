package database

import (
	"encoding/binary"
	"fmt"
	"os"
)

// Opens up database file a makes sure there is a metedata page
func Open() (*DB, error) {
	file, err := os.OpenFile(FileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open page: %w", err)
	}
	db, err := newDatabase(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}

func (DB *DB) Close() error {
	if err := DB.Pages.File.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// Adds to page if there is room to add the page
func (DB *DB) AddToPage(key string, value string) error {
	var pageID uint32
	// Get database metadata
	metadata, err := DB.Pages.readMetadata()
	if err != nil {
		return fmt.Errorf("failed to read : %w", err)
	}
	pageID = metadata.lastDataPage

	// Calculate required space
	dataSize := keyLenStorageSize + len(key) + valueLenStorageSize + len(value)
	pageMetadata, err := DB.Pages.readPageMetadata(pageID)
	if err != nil {
		return fmt.Errorf("failed to get metadata")
	}

	// Validate capacity
	freeSpace := int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	requiredSpace := slotSize + dataSize
	needsNewPage := freeSpace < requiredSpace
	if needsNewPage {
		pageMetadata, pageID, err = DB.Pages.ensureWritablePage()
		if err != nil {
			return fmt.Errorf("failed ensure writable data page: %w", err)
		}
		freeSpace = int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	}

	// Compute offsets
	slotOffset := getSlotOffset(pageMetadata)
	newDataOffset := getDataOffset(freeSpace, dataSize, pageMetadata.numEntries)

	// Compute buffers
	slotBuf := getSlotBuffer(uint16(newDataOffset), uint16(dataSize), slotNormal)
	dataBuf := getDataBuffer(key, value, dataSize)

	// Update file
	if err := DB.Pages.WriteBytes(slotBuf, slotOffset, pageID); err != nil {
		return fmt.Errorf("failed to write slot: %w", err)
	}
	if err := DB.Pages.WriteBytes(dataBuf, newDataOffset, pageID); err != nil {
		return fmt.Errorf("failed to write new data: %w", err)
	}

	// Add to b+tree
	DB.Insert(key, pageID, pageMetadata.numEntries)

	// Update page metadata
	pageMetadata.numEntries += 1
	pageMetadata.freeSpaceStart += uint16(slotSize)
	pageMetadata.freeSpaceEnd -= uint16(dataSize)
	if err := DB.Pages.updatePageMetadata(pageMetadata); err != nil {
		return fmt.Errorf("failed to update metadata")
	}

	return nil
}

// Check all the data records and if it matches the key updates the flag
func (DB *DB) Delete(key string) error {
	numOfPagesToRead, err := DB.Pages.getNumOfPagesToRead()
	if err != nil {
		return fmt.Errorf("failed to get page of pages to read")
	}
	// Loops over each page
	for pageID := 1; pageID <= numOfPagesToRead; pageID++ {
		// Check if data page
		pageMetadata, err := DB.Pages.readPageMetadata(uint32(pageID))
		if err != nil {
			return fmt.Errorf("failed to page %d: %w", pageID, err)
		}
		if pageMetadata.pageType != DataPage {
			continue
		}

		dataRecords, err := DB.Pages.read(uint32(pageID))
		if err != nil {
			return fmt.Errorf("failed to read page: %w", err)
		}
		for _, dataRecord := range dataRecords {
			if dataRecord.key != key {
				continue
			}
			// Change flag to deleted
			slotFlagBytes := make([]byte, slotFlagSize)
			binary.BigEndian.PutUint16(slotFlagBytes, slotDeleted)

			slotOffset := pageMetadataSize + (dataRecord.slotIndex * slotSize)
			slotFlagOffset := slotOffset + slotOffsetSize + slotLengthSize
			DB.Pages.WriteBytes(slotFlagBytes, slotFlagOffset, uint32(pageID))
		}
	}
	DB.T.Delete(key)
	return nil
}

// Return in a map of the key -> value. Remove duplicates and deleted records
func (DB *DB) SelectAll() (map[string]string, error) {
	data := map[string]string{}
	numOfPagesToRead, err := DB.Pages.getNumOfPagesToRead()
	if err != nil {
		return data, fmt.Errorf("failed to get page of pages to read")
	}
	// Loops over each page
	for pageID := 1; pageID <= numOfPagesToRead; pageID++ {
		dataRecords, err := DB.Pages.read(uint32(pageID))
		if err != nil {
			return data, fmt.Errorf("failed to format data: %w", err)
		}
		for _, dataRecord := range dataRecords {
			// Check  if value has been deleted
			if dataRecord.slot.flag != slotNormal {
				continue
			}
			data[dataRecord.key] = dataRecord.value
		}
	}
	return data, nil
}

// Uses b+tree to select the value returns; value, if found, error
func (DB *DB) Select(key string) (string, bool, error) {
	pageID, slotID, inTree, err := DB.FindKeyLocation(key)
	if err != nil {
		return "", false, fmt.Errorf("failed to select value: %w", err)
	}
	if inTree == false {
		return "", false, nil
	}
	slotBytes, err := DB.Pages.readSlot(pageID, slotID)
	if err != nil {
		return "", false, fmt.Errorf("failed to select value: %w", err)
	}
	formatedSlot := formatSlot(slotBytes)
	dataBytes, err := DB.Pages.readData(formatedSlot, pageID)
	if err != nil {
		return "", false, fmt.Errorf("failed to select value: %w", err)
	}
	_, value := readEntry(dataBytes)
	return value, true, nil
}
