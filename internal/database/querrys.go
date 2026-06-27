package database

import (
	"encoding/binary"
	"errors"
	"fmt"
)

type MathConditions string

const (
	GreaterThan          = ">"
	LessThan             = "<"
	GreaterThanOrEqualTo = ">="
	LessThanOrEqualTo    = "<="
)

type data struct {
	key   string
	value string
}

func newData(key string, value string) data {
	return data{
		key:   key,
		value: value,
	}
}

func Open() (*DB, error) {
	filename := "data/bubbly.db"
	return openDefault(filename)
}

func (db *DB) Close() error {
	if err := db.Pages.File.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// Adds to page if there is room to add the page
func (db *DB) AddToPage(key string, value string) error {
	var pageID uint32

	// Check input sizes
	if len(key) > maxKeySize && len(value) > maxValueSize {
		return errors.New("too long key and value")
	} else if len(key) > maxKeySize {
		return errors.New("too long key")
	} else if len(value) > maxValueSize {
		return errors.New("too long value")
	}

	// Get database metadata
	metadata, err := db.Pages.readMetadata()
	if err != nil {
		return fmt.Errorf("failed to read : %w", err)
	}
	pageID = metadata.lastDataPage

	// Calculate required space
	dataSize := keyLenStorageSize + len(key) + valueLenStorageSize + len(value)
	pageMetadata, err := db.Pages.readPageMetadata(pageID)
	if err != nil {
		return fmt.Errorf("failed to get metadata")
	}

	// Validate capacity
	freeSpace := int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	requiredSpace := slotSize + dataSize
	needsNewPage := freeSpace < requiredSpace
	if needsNewPage {
		pageMetadata, pageID, err = db.Pages.ensureWritablePage()
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
	if err := db.Pages.writeBytes(slotBuf, slotOffset, pageID); err != nil {
		return fmt.Errorf("failed to write slot: %w", err)
	}
	if err := db.Pages.writeBytes(dataBuf, newDataOffset, pageID); err != nil {
		return fmt.Errorf("failed to write new data: %w", err)
	}

	// Add to b+tree
	db.Insert(key, pageID, pageMetadata.numEntries)

	// Update page metadata
	pageMetadata.numEntries += 1
	pageMetadata.freeSpaceStart += uint16(slotSize)
	pageMetadata.freeSpaceEnd -= uint16(dataSize)
	if err := db.Pages.updatePageMetadata(pageMetadata); err != nil {
		return fmt.Errorf("failed to update metadata")
	}

	return nil
}

// Check all the data records and if it matches the key updates the flag
func (db *DB) Delete(key string) error {
	numOfPagesToRead, err := db.Pages.getNumOfPagesToRead()
	if err != nil {
		return fmt.Errorf("failed to get page of pages to read")
	}
	// Loops over each page
	for pageID := 1; pageID <= numOfPagesToRead; pageID++ {
		// Check if data page
		pageMetadata, err := db.Pages.readPageMetadata(uint32(pageID))
		if err != nil {
			return fmt.Errorf("failed to page %d: %w", pageID, err)
		}
		if pageMetadata.pageType != DataPage {
			continue
		}

		dataRecords, err := db.Pages.read(uint32(pageID))
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
			db.Pages.writeBytes(slotFlagBytes, slotFlagOffset, uint32(pageID))
		}
	}
	db.T.Delete(key)
	return nil
}

// Return in a map of the key -> value. Remove duplicates and deleted records
func (db *DB) SelectAll() (map[string]string, error) {
	data := map[string]string{}
	numOfPagesToRead, err := db.Pages.getNumOfPagesToRead()
	if err != nil {
		return data, fmt.Errorf("failed to get page of pages to read")
	}
	// Loops over each page
	for pageID := 1; pageID <= numOfPagesToRead; pageID++ {
		dataRecords, err := db.Pages.read(uint32(pageID))
		if errors.Is(err, ErrNotDataPage) {
			continue
		}
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
func (db *DB) Select(key string) (string, bool, error) {
	keyLocation, inTree, err := db.FindKeyLocation(key)
	if err != nil {
		return "", false, fmt.Errorf("failed to select value: %w", err)
	}
	if inTree == false {
		return "", false, nil
	}
	_, value, err := db.selectValue(keyLocation)
	if err != nil {
		return "", false, fmt.Errorf("failed to read value with key location: %w", err)
	}
	return value, true, nil
}

// Used like where x > 5.
func (db *DB) SelectWhere(condition MathConditions, boundaryKey string) (*[]data, error) {
	var selectedData *[]data
	var err error
	switch condition {
	case GreaterThan:
		selectedData, err = db.getMoreThan(boundaryKey, false)
		if err != nil {
			return nil, fmt.Errorf("failed to select data: %w", err)
		}
	case GreaterThanOrEqualTo:
		selectedData, err = db.getMoreThan(boundaryKey, true)
		if err != nil {
			return nil, fmt.Errorf("failed to select data: %w", err)
		}
	case LessThan:
		selectedData, err = db.getLessThan(boundaryKey, false)
		if err != nil {
			return nil, fmt.Errorf("failed to select data: %w", err)
		}
	case LessThanOrEqualTo:
		selectedData, err = db.getLessThan(boundaryKey, true)
		if err != nil {
			return nil, fmt.Errorf("failed to select data: %w", err)
		}
	default:
		return nil, errors.New("invaild condition")
	}
	return selectedData, nil
}
