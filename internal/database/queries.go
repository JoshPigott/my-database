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

func Open() (*DB, error) {
	filename := "data/bubbly.db"
	walFilename := "data/bubbly-wal.db"
	return openDefault(filename, walFilename)
}

func (db *DB) Close() error {
	if err := db.Pages.File.Close(); err != nil {
		return fmt.Errorf("failed to close database file: %w", err)
	}
	if err := db.Pages.walFile.Close(); err != nil {
		return fmt.Errorf("failed to close wal file: %w", err)
	}
	return nil
}

// Adds to page if there is room to add the page
func (db *DB) AddToPage(key string, value string) error {
	// Check input sizes
	if len(key) > maxKeySize && len(value) > maxValueSize {
		return errors.New("too long key and value")
	} else if len(key) > maxKeySize {
		return errors.New("too long key")
	} else if len(value) > maxValueSize {
		return errors.New("too long value")
	}

	// Get database metadata
	m, err := db.Pages.readDBMetadata()
	if err != nil {
		return fmt.Errorf("failed to read database metadata: %w", err)
	}
	pageID := m.lastDataPage

	// Calculate required space
	dataSize := keyLenStorageSize + len(key) + valueLenStorageSize + len(value)
	pm, err := db.Pages.readPageMetadata(pageID)
	if err != nil {
		return fmt.Errorf("failed to read pages metadata: %w", err)
	}

	// Validate capacity
	freeSpace := int(pm.freeSpaceEnd - pm.freeSpaceStart)
	requiredSpace := slotSize + dataSize
	needsNewPage := freeSpace < requiredSpace
	if needsNewPage {
		pm, pageID, err = db.Pages.ensureWritablePage()
		if err != nil {
			return fmt.Errorf("failed ensure writable data page: %w", err)
		}
		freeSpace = int(pm.freeSpaceEnd - pm.freeSpaceStart)
	}

	// Compute offsets
	slotOffset := getSlotOffset(pm)
	newDataOffset := getDataOffset(freeSpace, dataSize, pm.numEntries)

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
	if err := db.Insert(key, pageID, pm.numEntries); err != nil {
		return fmt.Errorf("failed insert into b+tree: %w", err)
	}

	// Update page metadata
	pm.numEntries += 1
	pm.freeSpaceStart += uint16(slotSize)
	pm.freeSpaceEnd -= uint16(dataSize)
	if err := db.Pages.updatePageMetadata(pm); err != nil {
		return fmt.Errorf("failed to update page metadata: %w", err)
	}
	if db.isManaulTX {
		return nil
	}
	if err := db.Pages.commit(); err != nil {
		return fmt.Errorf("failed to commit changes to database: %w", err)
	}
	return nil
}

// Check all the data records and if it matches the key updates the flag
func (db *DB) Delete(key string) error {
	numOfPagesToRead, err := db.Pages.getNumOfPagesToRead()
	if err != nil {
		return fmt.Errorf("failed to get page of pages to read: %w", err)
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
		records, err := db.Pages.readDataPage(uint32(pageID))
		if err != nil {
			return fmt.Errorf("failed to read page: %w", err)
		}
		// Checks if check is in page
		for _, record := range records {
			if record.data.key != key {
				continue
			}
			// Change flag to deleted
			slotFlagBytes := make([]byte, slotFlagSize)
			binary.BigEndian.PutUint16(slotFlagBytes, slotDeleted)

			slotOffset := pageMetadataSize + (record.slotIndex * slotSize)
			slotFlagOffset := slotOffset + slotOffsetSize + slotLengthSize
			db.Pages.writeBytes(slotFlagBytes, slotFlagOffset, uint32(pageID))
		}
	}
	if err := db.Root.Delete(key); err != nil {
		return fmt.Errorf("failed to to delete key in b+tree: %w", err)
	}
	if db.isManaulTX {
		return nil
	}
	if err := db.Pages.commit(); err != nil {
		return fmt.Errorf("failed to commit changes to database: %w", err)
	}
	return nil
}

// Return in a map of the key -> value. Remove duplicates and deleted records
func (db *DB) SelectAll() ([]Data, error) {
	data := []Data{}
	numOfPagesToRead, err := db.Pages.getNumOfPagesToRead()
	if err != nil {
		return []Data{}, fmt.Errorf("failed to get page of pages to read %w", err)
	}
	// Loops over each page
	for pageID := 1; pageID <= numOfPagesToRead; pageID++ {
		records, err := db.Pages.readDataPage(uint32(pageID))
		if errors.Is(err, ErrNotDataPage) {
			continue
		}
		if err != nil {
			return []Data{}, fmt.Errorf("failed to format data: %w", err)
		}

		for _, record := range records {
			// Check  if value has been deleted
			if record.slot.flag != slotNormal {
				continue
			}
			data = append(data, newData(record.data.key, record.data.value))
		}
	}
	return data, nil
}

// Uses b+tree to select the value returns; value, if found, error
func (db *DB) Select(key string) ([]Data, bool, error) {
	keyLocation, inTree, err := db.FindKeyLocation(key)
	if err != nil {
		return []Data{}, false, fmt.Errorf("failed to select value: %w", err)
	}
	if !inTree {
		return []Data{}, false, nil
	}
	data, err := db.Pages.selectData(keyLocation)
	if err != nil {
		return []Data{}, false, fmt.Errorf("failed to read value with key location: %w", err)
	}
	return []Data{data}, true, nil
}

// All you do to range querries selecting a range of data
func (db *DB) SelectWhere(condition MathConditions, boundaryKey string) ([]Data, error) {
	var selectedData []Data
	var err error
	switch condition {
	case GreaterThan:
		selectedData, err = db.getMoreThan(boundaryKey, false)
		if err != nil {
			return nil, fmt.Errorf("failed to select data more than %s: %w", boundaryKey, err)
		}
	case GreaterThanOrEqualTo:
		selectedData, err = db.getMoreThan(boundaryKey, true)
		if err != nil {
			return nil, fmt.Errorf("failed to select data more than or equal to %s: %w", boundaryKey, err)
		}
	case LessThan:
		selectedData, err = db.getLessThan(boundaryKey, false)
		if err != nil {
			return nil, fmt.Errorf("failed to select data less than %s: %w", boundaryKey, err)
		}
	case LessThanOrEqualTo:
		selectedData, err = db.getLessThan(boundaryKey, true)
		if err != nil {
			return nil, fmt.Errorf("failed to select data less than or equal to %s: %w", boundaryKey, err)
		}
	default:
		return nil, errors.New("invaild condition")
	}
	return selectedData, nil
}

// Clear existing WAL file and beings transaction
func (db *DB) BeginTX() error {
	err := db.Pages.needsReply()
	if err != nil {
		return fmt.Errorf("failed to reply WAL file to clear it: %w", err)
	}
	db.isManaulTX = true
	return nil
}

// Ends transaction and commits all changes to database
func (db *DB) EndTX() error {
	if err := db.Pages.commit(); err != nil {
		return fmt.Errorf("failed to commit to transaction: %w", err)
	}
	db.isManaulTX = false
	return nil
}
