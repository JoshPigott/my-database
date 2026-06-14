package database

import (
	"errors"
	"fmt"
	"io"
)

func (DB *DB) readPage(pageID uint32) ([]record, error) {
	_, _, dataRecords, err := DB.readFullPage(pageID)
	return dataRecords, err
}

// Read page to get the data in the page
func (DB *DB) readFullPage(pageID uint32) ([]byte, []slot, []record, error) {
	pageOffSet := getPageOffset(pageID)
	pageBytes := make([]byte, pageSize)
	if _, err := DB.File.Seek(pageOffSet, io.SeekStart); err != nil {
		return []byte{}, []slot{}, []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	if _, err := DB.File.Read(pageBytes); err != nil {
		return []byte{}, []slot{}, []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	pageMetadata := formatPageMetadata(pageBytes)
	slots := formatSlots(pageBytes, pageMetadata)
	dataRecords := readData(pageBytes, slots)
	return pageBytes, slots, dataRecords, nil
}

// Write new page at the end of the file
func (DB *DB) writePage(bytes []byte, pageType PageType) error {
	info, err := DB.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	if pageType == MetadataPage && info.Size() != 0 {
		return errors.New("failed to add metadata page: can't metadata not a top of database")
	}
	if _, err := DB.File.Seek(info.Size(), io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := DB.File.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := DB.File.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	return nil
}

// Open up database writes bytes and syncs bytes
func (DB *DB) WriteBytes(bytes []byte, offset int, pageID uint32) error {
	pageOffSet := getPageOffset(pageID)
	totalOffset := int64(offset) + pageOffSet

	// I want to read the bytes before
	if _, err := DB.File.Seek(totalOffset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := DB.File.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}

	// I want to read the bytes after
	if err := DB.File.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	return nil
}

// Reads page metadata (I will need to put what page later on)
func (DB *DB) readPageMetadata(pageID uint32) (pageMetadata, error) {
	pageOffSet := getPageOffset(pageID)
	metadataBytes := make([]byte, pageMetadataSize)
	if _, err := DB.File.Seek(pageOffSet, io.SeekStart); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read the page's metadata: %w", err)
	}
	if _, err := DB.File.Read(metadataBytes); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read page's metadata: %w", err)
	}
	pageMetadata := formatPageMetadata(metadataBytes)
	return pageMetadata, nil
}

// Reads metadata page and return the database metadata in a metadata struct
func (DB *DB) readMetadata() (metadata, error) {
	metadataBytes := make([]byte, metadataSize)
	if _, err := DB.File.Seek(pageMetadataSize, io.SeekStart); err != nil {
		return metadata{}, fmt.Errorf("failed to read metadata page: %w", err)
	}
	if _, err := DB.File.Read(metadataBytes); err != nil {
		return metadata{}, fmt.Errorf("failed to read metadata page: %w", err)
	}
	metadata := formatMetadata(metadataBytes)
	return metadata, nil
}

// Reads a particular slot
func (DB *DB) readSlot(pageID uint32, slotID uint16) ([]byte, error) {
	slotBytes := make([]byte, slotID)

	pageOffSet := getPageOffset(pageID)
	slotOffset := pageMetadataSize + (int64(slotID) * int64(slotSize))
	totalOffset := pageOffSet + slotOffset

	if _, err := DB.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	if _, err := DB.File.Read(slotBytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	return slotBytes, nil
}

// Gets bytes of a particular entry of data
func (DB *DB) readData(slot slot, pageID uint32) ([]byte, error) {
	dataBytes := make([]byte, slot.length)
	pageOffSet := getPageOffset(pageID)
	totalOffset := pageOffSet + int64(slot.offset)

	if _, err := DB.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	if _, err := DB.File.Read(dataBytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	return dataBytes, nil
}

func getPageOffset(pageID uint32) int64 {
	offset := (int64(pageID) - 1) * pageSize
	return offset
}

// // Notes this funcation is just for making my life easy with testing
// func beenCreated(numOfFiles int) (bool, error) {
// 	fileInfo, err := os.Stat(FileName)
// 	if errors.Is(err, os.ErrNotExist) {
// 		return true, nil
// 	}
// 	if err != nil {
// 		return false, fmt.Errorf("failed to check file info: %w", err)

// 	}
// 	if fileInfo.Size() == int64(numOfFiles*pageSize) {
// 		return true, nil
// 	}
// 	if fileInfo.Size() == 0 {
// 		return false, nil
// 	}
// 	return false, errors.New("failed to check if file has been created")
// }
