package database

import (
	"encoding/binary"
	"fmt"
	"os"
)

// This const is just for testing
const testPageID uint32 = 2

// THere are going to be querrys that the user can send

// Like think I should add in the next page

type DB struct {
	File *os.File
}

func newDatabase(file *os.File) DB {
	return DB{
		File: file,
	}
}

// Opens up database file a makes sure there is a metedata page
func Open() (*DB, error) {
	file, err := os.OpenFile(FileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open page: %w", err)
	}
	db := newDatabase(file)
	if err := db.ensureMetadataPage(); err != nil {
		return nil, fmt.Errorf("failed to create metadata page: %w", err)
	}
	if err := db.createFirstDataPage(); err != nil {
		return nil, fmt.Errorf("failed to create first data page")
	}
	return &db, nil
}

func (DB *DB) Close() error {
	if err := DB.File.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// Adds to page if there is room to add the page
func (DB *DB) AddToPage(key string, value string) error {
	var pageID uint32
	// Get database metadata
	metadata, err := DB.readMetadata()
	if err != nil {
		return fmt.Errorf("failed to read : %w", err)
	}
	// Get last used page
	if metadata.freePageStart != 0 {
		pageID = metadata.freePageStart - 1
	} else {
		pageID = metadata.totalNumPages
	}

	// Calculate required space
	dataSize := keyLengthStorageSize + len(key) + valueLengthStorageSize + len(value)
	pageMetadata, err := DB.readPageMetadata(pageID)
	if err != nil {
		return fmt.Errorf("failed to get metadata")
	}

	// Validate capacity
	freeSpace := int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	requiredSpace := slotSize + dataSize
	needsNewPage := freeSpace < requiredSpace
	if needsNewPage {
		pageMetadata, pageID, err = DB.ensureWritablePage(metadata, pageID)
		if err != nil {
			return fmt.Errorf("failed ensure writable data page: %w", err)
		}
		freeSpace = int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	}

	// Compute offsets and bytes to add
	slotOffset, newDataOffset := getOffsets(freeSpace, dataSize, pageMetadata.numslots)
	slotBuf := getSlotBuffer(uint16(newDataOffset), uint16(dataSize), slotNormal)
	dataBuf := getDataBuffer(key, value, dataSize)

	// Update file
	fmt.Println("Updating a slot")
	if err := DB.WriteBytes(slotBuf, slotOffset, pageID); err != nil {
		return fmt.Errorf("failed to write slot: %w", err)
	}
	fmt.Println("Updating a the data")
	if err := DB.WriteBytes(dataBuf, newDataOffset, pageID); err != nil {
		return fmt.Errorf("failed to write new data: %w", err)
	}
	if err := DB.updatePageMetadata(pageMetadata, dataSize); err != nil {
		return fmt.Errorf("failed to update metadata")
	}
	return nil
}

// Check all the data records and if it matches the key updates the flag
func (DB *DB) Delete(key string) error {
	numOfPagesToRead, err := DB.getNumOfPagesToRead()
	if err != nil {
		return fmt.Errorf("failed to get page of pages to read")
	}
	// Loops over each page
	for pageID := 1; pageID <= numOfPagesToRead; pageID++ {
		dataRecords, err := DB.readPage(uint32(pageID))
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
			DB.WriteBytes(slotFlagBytes, slotFlagOffset, uint32(pageID))
		}
	}
	return nil
}

// Return in a map of the key -> value. Remove duplicates and deleted records
func (DB *DB) SelectAll() (map[string]string, error) {
	data := map[string]string{}
	numOfPagesToRead, err := DB.getNumOfPagesToRead()
	if err != nil {
		return data, fmt.Errorf("failed to get page of pages to read")
	}
	// Loops over each page
	for pageID := 1; pageID <= numOfPagesToRead; pageID++ {
		dataRecords, err := DB.readPage(uint32(pageID))
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

// Trying to test out if DB write work on secound time
// I don't think it does so I want to what DB is
func (DB *DB) WriteSix() {
	fmt.Println("DB:", DB.File)
	if err := DB.WriteBytes([]byte{5}, 0, uint32(2)); err != nil {
		fmt.Println("failed to write 9")
	}
}

// Note this will chagne as I put B trees in
// func (DB *DB) Create() error {
// 	numOfFiles := 3
// 	created, err := beenCreated(numOfFiles)
// 	if err != nil {
// 		return err
// 	}
// 	if created == true {
// 		return nil
// 	}
// 	if err := DB.CreatePage(MetadataPage); err != nil {
// 		return err
// 	}
// 	// DB.CreatePage(RoutingPage)
// 	if err := DB.CreatePage(DataPage); err != nil {
// 		return err
// 	}
// 	return nil
// }
