package database

import (
	"errors"
	"fmt"
	"io"
)

// Read page to get the data in the page
func (DB *DB) readPage(pageID uint32) ([]record, error) {
	pageOffSet := getPageOffset(pageID)
	pageBytes := make([]byte, pageSize)
	if _, err := DB.File.Seek(pageOffSet, io.SeekStart); err != nil {
		return []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	if _, err := DB.File.Read(pageBytes); err != nil {
		return []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	pageMetadata := formatPageMetadata(pageBytes)
	slots := formatSlots(pageBytes, pageMetadata)
	dataRecords := readData(pageBytes, slots)
	// fmt.Println("dataRecords:", dataRecords)
	return dataRecords, nil
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
func (DB *DB) writeBytes(bytes []byte, offset int, pageID uint32) error {
	pageOffSet := getPageOffset(pageID)
	totalOffset := int64(offset) + pageOffSet

	if _, err := DB.File.Seek(totalOffset, io.SeekStart); err != nil {
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
