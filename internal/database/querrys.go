package database

import (
	"encoding/binary"
	"errors"
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
	// calculate required space
	dataSize := keyLengthStorageSize + len(key) + valueLengthStorageSize + len(value)
	pageMetadata, err := DB.readPageMetadata(testPageID)
	if err != nil {
		return fmt.Errorf("failed to get metadata")
	}

	// validate capacity
	freeSpace := int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	if freeSpace < (slotSize + dataSize) {
		return errors.New("failed to add page due to insufficient page storage capacity")
	}
	// compute offsets and bytes to add
	slotOffset, newDataOffset := getOffsets(freeSpace, dataSize, pageMetadata.numslots)
	slotBuf := getSlotBuffer(uint16(newDataOffset), uint16(dataSize), slotNormal)
	dataBuf := getDataBuffer(key, value, dataSize)

	// update file
	if err := DB.writeBytes(slotBuf, slotOffset, testPageID); err != nil {
		return fmt.Errorf("failed to write slot: %w", err)
	}
	if err := DB.writeBytes(dataBuf, newDataOffset, testPageID); err != nil {
		return fmt.Errorf("failed to write new data: %w", err)
	}
	if err := DB.updatePageMetadata(pageMetadata, dataSize); err != nil {
		return fmt.Errorf("failed to update metadata")
	}
	return nil
}

// Check all the data records and if it matches the key updates the flag
func (DB *DB) Delete(key string) error {
	dataRecords, err := DB.readPage(testPageID)
	if err != nil {
		return err
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
		DB.writeBytes(slotFlagBytes, slotFlagOffset, testPageID)
	}
	return nil
}

// Return in a map of the key -> value. Remove duplicates and deleted records
func (DB *DB) SelectAll() (map[string]string, error) {
	data := map[string]string{}
	dataRecords, err := DB.readPage(testPageID)
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
	return data, nil
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
