package database

import (
	"encoding/binary"
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
