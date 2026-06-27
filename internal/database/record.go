package database

import (
	"encoding/binary"
	"fmt"
	"strings"
)

type record struct {
	slotIndex int
	key       string
	value     string
	slot      slot
}

const (
	keyLenStorageSize   int = 2
	valueLenStorageSize int = 2
)

// Read a data entry and returns the key and value
func readEntry(dataBytes []byte) (string, string) {
	keyLength := getKeyLength(dataBytes)
	valueLength := getValueLength(dataBytes)
	key, valueStart := getKey(dataBytes, keyLength)
	value := getValue(dataBytes, valueStart, valueLength)
	return key, value
}

// Reads a page full of data
func readData(pageBytes []byte, slots []slot) []record {
	dataRecords := []record{}
	for i, slot := range slots {
		dataStart := slot.offset
		dataEnd := slot.length + dataStart
		dataBytes := pageBytes[dataStart:dataEnd]
		key, value := readEntry(dataBytes)
		dataRecord := record{
			slotIndex: i,
			key:       key,
			value:     value,
			slot:      slot,
		}
		dataRecords = append(dataRecords, dataRecord)
	}
	return dataRecords
}

// Gets key length from the page bytes
func getKeyLength(dataBytes []byte) int { // Right now I am not getting key length
	keyLengthBytes := dataBytes[0:keyLenStorageSize]
	keyLength := int(binary.BigEndian.Uint16(keyLengthBytes))
	return keyLength
}

// Get value length from the bytes
func getValueLength(dataBytes []byte) int {
	valueLengthEnd := keyLenStorageSize + valueLenStorageSize
	valueLenghtBytes := dataBytes[keyLenStorageSize:valueLengthEnd]
	valueLength := int(binary.BigEndian.Uint16(valueLenghtBytes))
	return valueLength
}

// Get key from the page bytes
func getKey(dataBytes []byte, keyLength int) (string, int) {
	keyStart := keyLenStorageSize + valueLenStorageSize
	keyEnd := keyStart + keyLength
	keyBytes := dataBytes[keyStart:keyEnd]
	key := string(keyBytes)
	return key, keyEnd
}

func getValue(dataBytes []byte, valueStart int, valueLength int) string {
	valueEnd := valueStart + valueLength
	valueBytes := dataBytes[valueStart:valueEnd]
	value := string(valueBytes)
	return value
}

// Gets page offset for the data
func getDataOffset(freeSpace int, dataSize int, numEntries uint16) int {
	totalSlotSize := int(numEntries) * slotSize
	totalSlotOffset := totalSlotSize + pageMetadataSize

	dataStart := totalSlotOffset + freeSpace
	newDataOffset := dataStart - dataSize
	return newDataOffset
}

// Get the value with a give key location
func (db *DB) selectValue(keyLocation *KeyLocation) (string, string, error) {
	slotBytes, err := db.Pages.readSlot(keyLocation.PageID, keyLocation.SlotID)
	if err != nil {
		return "", "", fmt.Errorf("failed to select value: %w", err)
	}
	formatedSlot := formatSlot(slotBytes)
	dataBytes, err := db.Pages.readData(formatedSlot, keyLocation.PageID)
	if err != nil {
		return "", "", fmt.Errorf("failed to select value: %w", err)
	}
	key, value := readEntry(dataBytes)
	return key, value, nil
}

// Used to slecet than or equal to a boundary key
func (DB *DB) getMoreThan(boundaryKey string, equalTo bool) (*[]data, error) {
	var selectedData []data
	n, _, err := DB.T.findNode(boundaryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to find leaf node: %w", err)
	}
	// Adds in speffic keys from the first node
	for i, key := range n.keys {
		if equalTo && strings.Compare(key, boundaryKey) >= 0 {
			key, value, err := DB.selectValue(n.keyLocations[i])
			if err != nil {
				return nil, fmt.Errorf("failed to get data from key location: %w", err)
			}
			data := newData(key, value)
			selectedData = append(selectedData, data)
		}
		if !equalTo && strings.Compare(key, boundaryKey) > 0 {
			key, value, err := DB.selectValue(n.keyLocations[i])
			if err != nil {
				return nil, fmt.Errorf("failed to get data from key location: %w", err)
			}
			data := newData(key, value)
			selectedData = append(selectedData, data)
		}
	}
	// Adds every key in the linked list until end
	for n.NextID != 0 {
		if n.Next != nil {
			n = n.Next
		} else {
			n, _ = n.pages.ReadPageNode(n.NextID)
		}
		for _, keyLocation := range n.keyLocations {
			key, value, err := DB.selectValue(keyLocation)
			if err != nil {
				return nil, fmt.Errorf("failed to get data from key location: %w", err)
			}
			data := newData(key, value)
			selectedData = append(selectedData, data)
		}
	}
	return &selectedData, nil
}

// Select all data less than a spefic boundary keys
func (DB *DB) getLessThan(boundaryKey string, equalTo bool) (*[]data, error) { // I need to the boundary key check
	var selectedData []data
	// Get left most node
	n := DB.T.root
	for !n.leaf {
		n = n.children[0]
	}
	for n != nil {
		for _, keyLocation := range n.keyLocations {
			key, value, err := DB.selectValue(keyLocation)
			if err != nil {
				return nil, fmt.Errorf("failed to get data from key location: %w", err)
			}
			if equalTo && strings.Compare(boundaryKey, key) == 0 {
				data := newData(key, value)
				selectedData = append(selectedData, data)
				return &selectedData, nil
			}
			if strings.Compare(boundaryKey, key) <= 0 {
				return &selectedData, nil
			}

			data := newData(key, value)
			selectedData = append(selectedData, data)

			if equalTo && strings.Compare(boundaryKey, key) <= 0 {
				return &selectedData, nil
			}
		}

		switch {
		case n.Next != nil:
			n = n.Next
		case n.NextID != 0:
			n, _ = n.pages.ReadPageNode(n.NextID)
		default:
			n = nil
		}
	}
	return &selectedData, nil
}
