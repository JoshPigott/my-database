package database

import (
	"errors"
	"fmt"
	"io"
)

func (Pages *Pages) read(pageID uint32) ([]record, error) {
	_, _, dataRecords, err := Pages.readFull(pageID)
	return dataRecords, err
}

// Read page to get the data in the page
func (Pages *Pages) readFull(pageID uint32) ([]byte, []slot, []record, error) {
	pageOffSet := getPageOffset(pageID)
	pageBytes := make([]byte, pageSize)
	if _, err := Pages.File.Seek(pageOffSet, io.SeekStart); err != nil {
		return []byte{}, []slot{}, []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	if _, err := Pages.File.Read(pageBytes); err != nil {
		return []byte{}, []slot{}, []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	pageMetadata := formatPageMetadata(pageBytes)
	slots := formatSlots(pageBytes, pageMetadata)
	dataRecords := readData(pageBytes, slots)
	return pageBytes, slots, dataRecords, nil
}

func (Pages *Pages) ReadBytes(size int, offset int, pageID uint32) ([]byte, error) {
	bytes := make([]byte, size)
	pageOffSet := getPageOffset(pageID)
	totalOffset := int64(offset) + pageOffSet
	if _, err := Pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := Pages.File.Read(bytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read file: %w", err)
	}
	return bytes, nil
}

// Write new page at the end of the file
func (Pages *Pages) write(bytes []byte, pageID uint32, pageType PageType) error {
	offset := getPageOffset(pageID)
	info, err := Pages.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
	}
	if pageType == MetadataPage && info.Size() != 0 {
		return errors.New("failed to add metadata page: can't metadata not a top of database")
	}
	if _, err := Pages.File.Seek(offset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := Pages.File.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := Pages.File.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	return nil
}

// Open up database writes bytes and syncs bytes
func (Pages *Pages) WriteBytes(bytes []byte, offset int, pageID uint32) error {
	pageOffSet := getPageOffset(pageID)
	totalOffset := int64(offset) + pageOffSet

	// I want to read the bytes before
	if _, err := Pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := Pages.File.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}

	// I want to read the bytes after
	if err := Pages.File.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	return nil
}

// Reads page metadata (I will need to put what page later on)
func (Pages *Pages) readPageMetadata(pageID uint32) (pageMetadata, error) {
	pageOffSet := getPageOffset(pageID)
	metadataBytes := make([]byte, pageMetadataSize)
	if _, err := Pages.File.Seek(pageOffSet, io.SeekStart); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read the page's metadata: %w", err)
	}
	if _, err := Pages.File.Read(metadataBytes); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read page's metadata: %w", err)
	}
	pageMetadata := formatPageMetadata(metadataBytes)
	return pageMetadata, nil
}

// Reads a particular slot
func (Pages *Pages) readSlot(pageID uint32, slotID uint16) ([]byte, error) {
	slotBytes := make([]byte, slotSize)

	pageOffSet := getPageOffset(pageID)
	slotOffset := pageMetadataSize + (int64(slotID) * int64(slotSize))
	totalOffset := pageOffSet + slotOffset

	if _, err := Pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	if _, err := Pages.File.Read(slotBytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	return slotBytes, nil
}

// Gets bytes of a particular entry of data
func (Pages *Pages) readData(slot slot, pageID uint32) ([]byte, error) {
	dataBytes := make([]byte, slot.length)
	pageOffSet := getPageOffset(pageID)
	totalOffset := pageOffSet + int64(slot.offset)

	if _, err := Pages.File.Seek(totalOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	if _, err := Pages.File.Read(dataBytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read slot: %w", err)
	}
	return dataBytes, nil
}

func getPageOffset(pageID uint32) int64 {
	offset := (int64(pageID) - 1) * pageSize
	return offset
}
