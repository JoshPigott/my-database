package pages

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// !!!!! NOTE THIS ONLY WORKS FOR ONE PAGE ATM !!!!!

// My main goal is to be able to create, write, read pages

// THing I will need to use are Seek() + Write()
// 4kb pages 4096 bytes

/*
Notes about page structure
So the slots part keeps track of where each key value thing starts.
The end will just be the start of the one before exclusive
Then in the data itself there will the key len and the value len
Both the key and the value lenght are two bytes each

Slot lenght will be two bytes longer and beign right ater the header and
where the data starts
*/

type PageType int8

const (
	pageFileName = "data/database-page.db"
	pageSize     = 4096

	MetadataPage PageType = 0
	RootPage     PageType = 1
	RoutingPage  PageType = 2
	DataPage     PageType = 3

	metadataSize        = 7
	pageTypeIndex       = 0
	numslotsIndex       = 1
	freeSpaceStartIndex = 3
	freeSpaceEndIndex   = 5

	slotSize               int = 2
	keyLengthStorageSize   int = 2
	valueLengthStorageSize int = 2

	defaultNumSlots uint16 = 0
)

type pageMetadata struct {
	pageType       PageType
	numslots       uint16
	freeSpaceStart uint16
	freeSpaceEnd   uint16
}

// Formats bytes into a metadata struc
func formatPageMetadata(metadataBytes []byte) pageMetadata {
	return pageMetadata{
		pageType:       PageType(int8(metadataBytes[pageTypeIndex])),
		numslots:       binary.BigEndian.Uint16(metadataBytes[numslotsIndex:freeSpaceStartIndex]),
		freeSpaceStart: binary.BigEndian.Uint16(metadataBytes[freeSpaceStartIndex:freeSpaceEndIndex]),
		freeSpaceEnd:   binary.BigEndian.Uint16(metadataBytes[freeSpaceEndIndex:metadataSize]),
	}
}

// Create an empty page
func CreatePage(pageType PageType) error {
	bytes := make([]byte, pageSize)
	pageMetadata := pageMetadata{
		pageType:       pageType,
		numslots:       defaultNumSlots,
		freeSpaceStart: metadataSize,
		freeSpaceEnd:   pageSize,
	}
	buf := createMetadataBuffer(pageMetadata)
	copy(bytes, buf)
	err := writePage(bytes)
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	return nil
}

func AddToPage(key string, value string) error { // key string, value string
	dataSize := keyLengthStorageSize + len(key) + valueLengthStorageSize + len(value)
	pageMetadata, err := readMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata")
	}
	// Checks if is free space in the page to add data
	freeSpace := int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	if freeSpace < (slotSize + dataSize) {
		return errors.New("failed to add page due to insufficient page storage capacity")
	}
	slotOffset, newDataOffset := getOffsets(freeSpace, dataSize, pageMetadata.numslots)
	slotBuf := getSlotBuffer(newDataOffset)
	dataBuf := getDataBuffer(key, value, dataSize)

	if err := writeBytes(slotBuf, slotOffset); err != nil {
		return fmt.Errorf("failed to write slot: %w", err)
	}
	if err := writeBytes(dataBuf, newDataOffset); err != nil {
		return fmt.Errorf("failed to write new data: %w", err)
	}
	if err := updateMetadata(pageMetadata, dataSize); err != nil {
		return fmt.Errorf("failed to update metadata")
	}
	return nil
}

func deleteFromPage() {

}

func ReadPage() (map[string]string, error) {
	pageBytes := make([]byte, pageSize)
	pageFile, err := os.OpenFile(pageFileName, os.O_RDONLY, 0644)
	if err != nil {
		return map[string]string{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer pageFile.Close()
	if _, err = pageFile.Seek(0, io.SeekStart); err != nil {
		return map[string]string{}, fmt.Errorf("failed to read file: %w", err)
	}
	if _, err := pageFile.Read(pageBytes); err != nil {
		return map[string]string{}, fmt.Errorf("failed to read file: %w", err)
	}
	pageMetadata := formatPageMetadata(pageBytes)
	slots := formatSlots(pageBytes, pageMetadata)
	data := readData(pageBytes, slots)
	return data, nil
}

// Iterates over each slot connverting bytes to make a slice
func formatSlots(pageBytes []byte, pageMetadata pageMetadata) []int {
	fmt.Println("pageMetadata.numslots:", pageMetadata.numslots)
	var formatedSlots []int
	slotSize := int(pageMetadata.numslots) * slotSize
	slotsBytes := pageBytes[metadataSize:(metadataSize + slotSize)]
	for i := range int(pageMetadata.numslots) {
		slotBtyes := slotsBytes[(i * 2) : (i*2)+slotSize]
		slot := int(binary.BigEndian.Uint16(slotBtyes))
		formatedSlots = append(formatedSlots, slot)
	}
	return formatedSlots
}

func readData(pageBytes []byte, slots []int) map[string]string {
	data := map[string]string{}
	for _, slot := range slots {
		keyLength, valueLengthStart := getKeyLength(pageBytes, slot)
		valueLength, keyStart := getValueLength(pageBytes, valueLengthStart)
		key, valueStart := getKey(pageBytes, keyStart, keyLength)
		value := getValue(pageBytes, valueStart, valueLength)
		data[key] = value
	}
	return data
}

// Create the metadata buffer
func createMetadataBuffer(pageMetadata pageMetadata) []byte {
	buf := make([]byte, metadataSize)
	buf[pageTypeIndex] = byte(pageMetadata.pageType)
	binary.BigEndian.PutUint16(buf[numslotsIndex:freeSpaceStartIndex], pageMetadata.numslots)
	binary.BigEndian.PutUint16(buf[freeSpaceStartIndex:freeSpaceEndIndex], pageMetadata.freeSpaceStart)
	binary.BigEndian.PutUint16(buf[freeSpaceEndIndex:metadataSize], pageMetadata.freeSpaceEnd)
	return buf
}

// Writes a page file and sync it insuring bytes have written
// Note this is only used to create the first page
func writePage(bytes []byte) error {
	pageFile, err := os.OpenFile(pageFileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("failed to open page: %w", err)
	}
	// fmt.Println("bytes before:", bytes)
	if _, err := pageFile.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := pageFile.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := pageFile.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// Open up database writes and syncs bytes
func writeBytes(bytes []byte, offset int) error {
	pageFile, err := os.OpenFile(pageFileName, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open page: %w", err)
	}
	if _, err = pageFile.Seek(int64(offset), io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := pageFile.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := pageFile.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := pageFile.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}

// Read a page metadata (I will need to put what page later on)
func readMetadata() (pageMetadata, error) {
	metadataBytes := make([]byte, metadataSize)
	pageFile, err := os.OpenFile(pageFileName, os.O_RDONLY, 0644)
	if err != nil {
		return pageMetadata{}, fmt.Errorf("failed to open file: %w", err)
	}
	defer pageFile.Close()
	if _, err = pageFile.Seek(0, io.SeekStart); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read metadata: %w", err)
	}
	if _, err := pageFile.Read(metadataBytes); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read metadata: %w", err)
	}
	pageMetadata := formatPageMetadata(metadataBytes)
	return pageMetadata, err
}

// Rewrite page metadata
func updateMetadata(oldPageMetadata pageMetadata, dataSize int) error {
	newPageMetadata := pageMetadata{
		pageType:       oldPageMetadata.pageType,
		numslots:       oldPageMetadata.numslots + 1,
		freeSpaceStart: oldPageMetadata.freeSpaceStart + uint16(slotSize),
		freeSpaceEnd:   oldPageMetadata.freeSpaceEnd - uint16(dataSize),
	}
	buf := createMetadataBuffer(newPageMetadata)
	err := writeBytes(buf, 0)
	return err
}

// Finds the data offset
func getOffsets(freeSpace int, dataSize int, numslots uint16) (int, int) {
	totalSlotSize := int(numslots) * slotSize
	slotOffset := totalSlotSize + metadataSize

	dataStart := slotOffset + freeSpace
	newDataOffset := dataStart - dataSize
	return slotOffset, newDataOffset
}

// and convert that into bytes and return it
func getSlotBuffer(newDataOffset int) []byte {
	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(newDataOffset))
	return buf
}

// Writes the len and value of key and the value into a buffer
func getDataBuffer(key string, value string, dataSize int) []byte {
	buf := make([]byte, dataSize)
	var keyLength uint16 = uint16(len(key))
	var valueLength uint16 = uint16(len(value))

	binary.BigEndian.PutUint16(buf[0:2], keyLength)
	binary.BigEndian.PutUint16(buf[2:4], valueLength)

	valueStart := 4 + len(key)
	copy(buf[4:], []byte(key))
	copy(buf[valueStart:], []byte(value))
	return buf
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
