package database

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
	FileName = "data/bubbly.db"
	pageSize = 4096

	MetadataPage PageType = 0
	RootPage     PageType = 1
	RoutingPage  PageType = 2
	DataPage     PageType = 3

	metadataSize        = 7
	pageTypeIndex       = 0
	numslotsIndex       = 1
	freeSpaceStartIndex = 3
	freeSpaceEndIndex   = 5

	// Slot info
	slotSize        int    = 6
	slotOffsetSize  int    = 2
	slotLengthSize  int    = 2
	slotFlagSize    int    = 2
	slotOffsetIndex int    = 0
	slotLengthIndex int    = 2
	slotFlagIndex   int    = 4
	slotNormal      uint16 = 0
	slotDeleted     uint16 = 1 << 0

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

type slot struct {
	offset uint16
	length uint16
	flag   uint16
}

type record struct {
	slotIndex int
	key       string
	value     string
	slot      slot
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

// Format bytes into a slot struc
func formatSlot(slotBytes []byte) slot {
	return slot{
		offset: binary.BigEndian.Uint16(slotBytes[slotOffsetIndex:slotLengthIndex]),
		length: binary.BigEndian.Uint16(slotBytes[slotLengthIndex:slotFlagIndex]),
		flag:   binary.BigEndian.Uint16(slotBytes[slotFlagIndex:slotSize]),
	}
}

// Note this will chagne as I put B trees in
// func (DB *DB) Create() error {
// 	numOfFiles := 3
// 	created, err := beenCreated(numOfFiles)
// 	if err != nil {
// 		return err
// 	}
// 	if created == true {
// 		return nil
// 	}
// 	DB.CreatePage(MetadataPage)
// 	DB.CreatePage(RoutingPage)
// 	DB.CreatePage(DataPage)
// 	return nil
// }

// Create an empty page
func (DB *DB) CreatePage(pageType PageType) error {
	bytes := make([]byte, pageSize)
	pageMetadata := pageMetadata{
		pageType:       pageType,
		numslots:       defaultNumSlots,
		freeSpaceStart: metadataSize,
		freeSpaceEnd:   pageSize,
	}
	buf := createMetadataBuffer(pageMetadata)
	copy(bytes, buf)
	err := DB.writePage(bytes)
	if err != nil {
		return fmt.Errorf("failed to create page: %w", err)
	}
	return nil
}

// Adds to page if there is room to add the page
func (DB *DB) AddToPage(key string, value string) error {
	// calculate required space
	dataSize := keyLengthStorageSize + len(key) + valueLengthStorageSize + len(value)
	pageMetadata, err := DB.readMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata")
	}
	// validate capacity
	freeSpace := int(pageMetadata.freeSpaceEnd - pageMetadata.freeSpaceStart)
	if freeSpace < (slotSize + dataSize) {
		return errors.New("failed to add page due to insufficient page storage capacity")
	}
	// compute offsets and bytes to add
	slotOffset, newDataOffset := getOffsets(freeSpace, dataSize, pageMetadata.numslots)
	slotBuf := getSlotBuffer(uint16(newDataOffset), uint16(dataSize), slotNormal)
	dataBuf := getDataBuffer(key, value, dataSize)

	// update file
	if err := DB.writeBytes(slotBuf, slotOffset); err != nil {
		return fmt.Errorf("failed to write slot: %w", err)
	}
	if err := DB.writeBytes(dataBuf, newDataOffset); err != nil {
		return fmt.Errorf("failed to write new data: %w", err)
	}
	if err := DB.updateMetadata(pageMetadata, dataSize); err != nil {
		return fmt.Errorf("failed to update metadata")
	}
	return nil
}

// Check all the data records and if it matches the key updates the flag
func (DB *DB) Delete(key string) error {
	dataRecords, err := DB.readPage()
	if err != nil {
		return err
	}
	for _, dataRecord := range dataRecords {
		if dataRecord.key != key {
			continue
		}
		// Change flag to deleted
		slotFlagBytes := make([]byte, slotFlagSize)
		binary.BigEndian.PutUint16(slotFlagBytes, slotDeleted)

		slotOffset := metadataSize + (dataRecord.slotIndex * slotSize)
		slotFlagOffset := slotOffset + slotOffsetSize + slotLengthSize
		DB.writeBytes(slotFlagBytes, slotFlagOffset)
	}
	return nil
}

// Return in a map of the key -> value. Remove duplicates and deleted records
func (DB *DB) SelectAll() (map[string]string, error) {
	data := map[string]string{}
	dataRecords, err := DB.readPage()
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
	return data, nil
}

// Read page to get the data in the page
func (DB *DB) readPage() ([]record, error) {
	pageBytes := make([]byte, pageSize)
	if _, err := DB.File.Seek(0, io.SeekStart); err != nil {
		return []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	if _, err := DB.File.Read(pageBytes); err != nil {
		return []record{}, fmt.Errorf("failed to read file: %w", err)
	}
	pageMetadata := formatPageMetadata(pageBytes)
	slots := formatSlots(pageBytes, pageMetadata)
	dataRecords := readData(pageBytes, slots)
	return dataRecords, nil
}

// Iterates over each slot connverting bytes to make a slice
func formatSlots(pageBytes []byte, pageMetadata pageMetadata) []slot {
	var formatedSlots []slot
	totalSlotSize := int(pageMetadata.numslots) * slotSize
	slotsBytes := pageBytes[metadataSize:(metadataSize + totalSlotSize)]
	for i := range int(pageMetadata.numslots) {
		slotBtyes := slotsBytes[(i * slotSize) : (i*slotSize)+slotSize]
		slot := formatSlot(slotBtyes)
		formatedSlots = append(formatedSlots, slot)
	}
	return formatedSlots
}

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

// Create the metadata buffer
func createMetadataBuffer(pageMetadata pageMetadata) []byte {
	buf := make([]byte, metadataSize)
	buf[pageTypeIndex] = byte(pageMetadata.pageType)
	binary.BigEndian.PutUint16(buf[numslotsIndex:freeSpaceStartIndex], pageMetadata.numslots)
	binary.BigEndian.PutUint16(buf[freeSpaceStartIndex:freeSpaceEndIndex], pageMetadata.freeSpaceStart)
	binary.BigEndian.PutUint16(buf[freeSpaceEndIndex:metadataSize], pageMetadata.freeSpaceEnd)
	return buf
}

// Write new page at the end of the file
func (DB *DB) writePage(bytes []byte) error {
	info, err := DB.File.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file size: %w", err)
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
func (DB *DB) writeBytes(bytes []byte, offset int) error {
	if _, err := DB.File.Seek(int64(offset), io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := DB.File.Write(bytes); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}

	if err := DB.File.Sync(); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if _, err := DB.File.Seek(int64(0), io.SeekStart); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	return nil
}

// Read a page metadata (I will need to put what page later on)
func (DB *DB) readMetadata() (pageMetadata, error) {
	metadataBytes := make([]byte, metadataSize)
	if _, err := DB.File.Seek(0, io.SeekStart); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read metadata: %w", err)
	}
	if _, err := DB.File.Read(metadataBytes); err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read metadata: %w", err)
	}
	pageMetadata := formatPageMetadata(metadataBytes)
	return pageMetadata, nil
}

// Rewrite page metadata
func (DB *DB) updateMetadata(oldPageMetadata pageMetadata, dataSize int) error {
	newPageMetadata := pageMetadata{
		pageType:       oldPageMetadata.pageType,
		numslots:       oldPageMetadata.numslots + 1,
		freeSpaceStart: oldPageMetadata.freeSpaceStart + uint16(slotSize),
		freeSpaceEnd:   oldPageMetadata.freeSpaceEnd - uint16(dataSize),
	}
	buf := createMetadataBuffer(newPageMetadata)
	err := DB.writeBytes(buf, 0)
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
func getSlotBuffer(newDataOffset uint16, dataLength uint16, flag uint16) []byte {
	buf := make([]byte, slotSize)
	binary.BigEndian.PutUint16(buf[slotOffsetIndex:slotOffsetIndex+slotOffsetSize], newDataOffset)
	binary.BigEndian.PutUint16(buf[slotLengthIndex:slotLengthIndex+slotLengthSize], dataLength)
	binary.BigEndian.PutUint16(buf[slotFlagIndex:slotFlagIndex+slotFlagSize], flag)
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

func beenCreated(numOfFiles int) (bool, error) {
	fileInfo, err := os.Stat(FileName)
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check file info: %w", err)

	}
	if fileInfo.Size() == int64(numOfFiles*pageSize) {
		return true, nil
	}
	if fileInfo.Size() == 0 {
		return false, nil
	}
	return false, errors.New("failed to check if file has been created")
}
