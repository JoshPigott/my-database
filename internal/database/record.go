package database

import (
	"encoding/binary"
	"fmt"
	"strings"
)

const (
	keyLenStorageSize   int = 2
	valueLenStorageSize int = 2
)

type record struct {
	slotIndex int
	data      data
	slot      slot
}

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

// Read a data entry and returns the key and value
func readEntry(dataBytes []byte) data {
	keyLength := getKeyLength(dataBytes)
	valueLength := getValueLength(dataBytes)
	key, valueStart := getKey(dataBytes, keyLength)
	value := getValue(dataBytes, valueStart, valueLength)
	return newData(key, value)
}

// Reads a page full of data
func formatData(pageBytes []byte, slots []slot) []record {
	dataRecords := []record{}
	for i, slot := range slots {
		dataStart := slot.offset
		dataEnd := slot.length + dataStart
		dataBytes := pageBytes[dataStart:dataEnd]
		dataRecord := record{
			slotIndex: i,
			data:      readEntry(dataBytes),
			slot:      slot,
		}
		dataRecords = append(dataRecords, dataRecord)
	}
	return dataRecords
}

// Gets key length from the page bytes
func getKeyLength(dataBytes []byte) int {
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
func (pages *Pages) selectValue(keyLocation *KeyLocation) (string, string, error) {
	// get slot
	slotOffset := pageMetadataSize + (int(keyLocation.SlotID) * slotSize)
	slotBytes, err := pages.ReadBytes(slotSize, slotOffset, keyLocation.PageID)
	if err != nil {
		return "", "", fmt.Errorf("failed to read slot: %w", err)
	}
	formatedSlot := formatSlot(slotBytes)

	// Get data
	dataBytes, err := pages.ReadBytes(int(formatedSlot.length), int(formatedSlot.offset), keyLocation.PageID)
	if err != nil {
		return "", "", fmt.Errorf("failed to read data: %w", err)
	}
	data := readEntry(dataBytes)
	return data.key, data.value, nil
}

// Used to slecet than or equal to a boundary key
func (db *DB) getMoreThan(boundaryKey string, equalTo bool) ([]data, error) {
	var selectedData []data
	n, _, err := db.Root.findNode(boundaryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to find node: %w", err)
	}
	// Adds in speffic keys from the first node
	for i, key := range n.keys {
		if equalTo && strings.Compare(key, boundaryKey) >= 0 {
			key, value, err := db.Pages.selectValue(n.keyLocations[i])
			if err != nil {
				return nil, fmt.Errorf("failed to select value: %w", err)
			}
			data := newData(key, value)
			selectedData = append(selectedData, data)
		}
		if !equalTo && strings.Compare(key, boundaryKey) > 0 {
			key, value, err := db.Pages.selectValue(n.keyLocations[i])
			if err != nil {
				return nil, fmt.Errorf("failed to select value: %w", err)
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
			n, err = n.pages.ReadPageNode(n.NextID)
			if err != nil {
				return nil, fmt.Errorf("failed to read next node page: %w", err)
			}
		}
		for _, keyLocation := range n.keyLocations {
			key, value, err := db.Pages.selectValue(keyLocation)
			if err != nil {
				return nil, fmt.Errorf("failed to select value: %w", err)
			}
			data := newData(key, value)
			selectedData = append(selectedData, data)
		}
	}
	return selectedData, nil
}

// Select all data less than a spefic boundary keys
func (db *DB) getLessThan(boundaryKey string, equalTo bool) ([]data, error) {
	var selectedData []data
	var err error
	// Get left most node
	n := db.Root
	for !n.leaf {
		if n.children[0] == nil {
			n, err = n.pages.ReadPageNode(n.childPageIDs[0])
			if err != nil {
				return []data{}, fmt.Errorf("failed to read left child node page: %w", err)
			}
		}
		n = n.children[0]

	}
	for n != nil {
		nodesData, err := n.selectNodesData(boundaryKey, equalTo)
		if err != nil {
			return nil, fmt.Errorf("failed to select current nodes data: %w", err)
		}
		selectedData = append(selectedData, nodesData...)

		switch {
		case n.Next != nil:
			n = n.Next
		case n.NextID != 0:
			n, err = n.pages.ReadPageNode(n.NextID)
			if err != nil {
				return nil, fmt.Errorf("failed to read next nodes page: %w", err)
			}
		default:
			n = nil
		}
	}
	return selectedData, nil
}

// Read current node keys and selects data less the boundary key
func (n *node) selectNodesData(boundaryKey string, equalTo bool) ([]data, error) {
	var nodesData []data
	for i, key := range n.keys {
		if strings.Compare(boundaryKey, key) < 0 {
			return nodesData, nil
		}
		if equalTo && strings.Compare(boundaryKey, key) == 0 {
			key, value, err := n.pages.selectValue(n.keyLocations[i])
			if err != nil {
				return nil, fmt.Errorf("failed to select valu: %w", err)
			}
			data := newData(key, value)
			nodesData = append(nodesData, data)
			return nodesData, nil
		}
		if strings.Compare(boundaryKey, key) > 0 {
			key, value, err := n.pages.selectValue(n.keyLocations[i])
			if err != nil {
				return nil, fmt.Errorf("failed to select value: %w", err)
			}
			data := newData(key, value)
			nodesData = append(nodesData, data)
		}
	}
	return nodesData, nil
}
