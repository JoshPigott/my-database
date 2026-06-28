package database

import (
	"errors"
	"fmt"
	"io"
)

var ErrNotDataPage = errors.New("failed to read page due to not data page")

func (pages *Pages) read(pageID uint32) ([]record, error) {
	_, _, dataRecords, err := pages.readFull(pageID)
	return dataRecords, err
}

// Read page to get the data in a data page
func (pages *Pages) readFull(pageID uint32) ([]byte, []slot, []record, error) { // Ugly
	pageOffSet := getPageOffset(pageID)
	pageBytes := make([]byte, pageSize)
	if _, err := pages.File.Seek(pageOffSet, io.SeekStart); err != nil {
		return []byte{}, []slot{}, []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	if _, err := pages.File.Read(pageBytes); err != nil {
		return []byte{}, []slot{}, []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	pageMetadata := formatPageMetadata(pageBytes)
	if pageMetadata.pageType != DataPage {
		return []byte{}, []slot{}, []record{}, ErrNotDataPage
	}
	slots := formatSlots(pageBytes, pageMetadata)
	dataRecords := readData(pageBytes, slots)
	return pageBytes, slots, dataRecords, nil
}

func (pages *Pages) ReadBytes(size int, offset int, pageID uint32) ([]byte, error) {
	bytes := make([]byte, size)
	pageOffSet := getPageOffset(pageID)
	totalOffset := int64(offset) + pageOffSet
	if _, err := pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := pages.File.Read(bytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes, nil
}

// Write new page at the end of the file
func (pages *Pages) write(bytes []byte, pageID uint32, pageType PageType) error {
	offset := getPageOffset(pageID)
	info, err := pages.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	if pageType == MetadataPage && info.Size() != 0 {
		return errors.New("failed to add metadata page: can't metadata not a top of database")
	}
	if _, err := pages.File.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := pages.File.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := pages.File.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	return nil
}

// Open up database writes bytes and syncs bytes
func (pages *Pages) writeBytes(bytes []byte, offset int, pageID uint32) error {
	pageOffSet := getPageOffset(pageID)
	totalOffset := int64(offset) + pageOffSet

	// I want to read the bytes before
	if _, err := pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := pages.File.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}

	// I want to read the bytes after
	if err := pages.File.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	return nil
}

// Reads page metadata (I will need to put what page later on)
func (pages *Pages) readPageMetadata(pageID uint32) (pageMetadata, error) {
	pageOffSet := getPageOffset(pageID)
	metadataBytes := make([]byte, pageMetadataSize)
	if _, err := pages.File.Seek(pageOffSet, io.SeekStart); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read the page's metadata: %w", err)
	}
	if _, err := pages.File.Read(metadataBytes); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read page's metadata: %w", err)
	}
	pageMetadata := formatPageMetadata(metadataBytes)
	return pageMetadata, nil
}

// Reads a particular slot
func (pages *Pages) readSlot(pageID uint32, slotID uint16) ([]byte, error) {
	slotBytes := make([]byte, slotSize)

	pageOffSet := getPageOffset(pageID)
	slotOffset := pageMetadataSize + (int64(slotID) * int64(slotSize))
	totalOffset := pageOffSet + slotOffset

	if _, err := pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	if _, err := pages.File.Read(slotBytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	return slotBytes, nil
}

// Gets bytes of a particular entry of data
func (pages *Pages) readData(slot slot, pageID uint32) ([]byte, error) {
	dataBytes := make([]byte, slot.length)
	pageOffSet := getPageOffset(pageID)
	totalOffset := pageOffSet + int64(slot.offset)

	if _, err := pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	if _, err := pages.File.Read(dataBytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	return dataBytes, nil
}

func getPageOffset(pageID uint32) int64 {
	offset := (int64(pageID) - 1) * pageSize
	return offset
}
