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
	keyLengthStorageSize   int = 2
	valueLengthStorageSize int = 2
)

func readData(pageBytes []byte, slots []slot) []record {
	dataRecords := []record{}
	for i, slot := range slots {
		keyLength, valueLengthStart := getKeyLength(pageBytes, int(slot.offset))
		valueLength, keyStart := getValueLength(pageBytes, valueLengthStart)
		key, valueStart := getKey(pageBytes, keyStart, keyLength)
		value := getValue(pageBytes, valueStart, valueLength)
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
func getKeyLength(pageBytes []byte, slot int) (int, int) {
	keyLengthStart := slot
	keyLengthEnd := slot + keyLengthStorageSize
	keyLengthBytes := pageBytes[keyLengthStart:keyLengthEnd]
	keyLength := int(binary.BigEndian.Uint16(keyLengthBytes))
	return keyLength, keyLengthEnd
}

// Get value length from the bytes
func getValueLength(pageBytes []byte, valueLengthStart int) (int, int) {
	valueLengthEnd := valueLengthStart + valueLengthStorageSize
	valueLenghtBytes := pageBytes[valueLengthStart:valueLengthEnd]
	valueLength := int(binary.BigEndian.Uint16(valueLenghtBytes))
	return valueLength, valueLengthEnd
}

// Get key from the page bytes
func getKey(pageBytes []byte, keyStart int, keyLength int) (string, int) {
	keyEnd := keyStart + keyLength
	keyBytes := pageBytes[keyStart:keyEnd]
	key := string(keyBytes)
	return key, keyEnd
}

func getValue(pageBytes []byte, valueStart int, valueLength int) string {
	valueEnd := valueStart + valueLength
	valueBytes := pageBytes[valueStart:valueEnd]
	value := string(valueBytes)
	return value
}

// Gets page offset for the data
func getDataOffset(freeSpace int, dataSize int, numslots uint16) int {
	totalSlotSize := int(numslots) * slotSize
	totalSlotOffset := totalSlotSize + pageMetadataSize

	dataStart := totalSlotOffset + freeSpace
	newDataOffset := dataStart - dataSize
	return newDataOffset
}
