package database

import (
	"encoding/binary"
	"strings"
)

const (
	slotSize        int    = 6
	slotOffsetSize  int    = 2
	slotLengthSize  int    = 2
	slotFlagSize    int    = 2
	slotOffsetIndex int    = 0
	slotLengthIndex int    = 2
	slotFlagIndex   int    = 4
	slotNormal      uint16 = 0
	slotDeleted     uint16 = 1 << 0
)

type slot struct {
	offset uint16
	length uint16
	flag   uint16
}

// Format bytes into a slot struc
func formatSlot(slotBytes []byte) slot {
	return slot{
		offset: binary.BigEndian.Uint16(slotBytes[slotOffsetIndex:slotLengthIndex]),
		length: binary.BigEndian.Uint16(slotBytes[slotLengthIndex:slotFlagIndex]),
		flag:   binary.BigEndian.Uint16(slotBytes[slotFlagIndex:slotSize]),
	}
}

// Iterates over each slot connverting bytes to make a slice
func formatSlots(pageBytes []byte, pageMetadata pageMetadata) []slot {
	var formatedSlots []slot
	totalSlotSize := int(pageMetadata.numSlots) * slotSize
	slotsBytes := pageBytes[pageMetadataSize:(pageMetadataSize + totalSlotSize)]
	for i := range int(pageMetadata.numSlots) {
		slotBtyes := slotsBytes[(i * slotSize) : (i*slotSize)+slotSize]
		slot := formatSlot(slotBtyes)
		formatedSlots = append(formatedSlots, slot)
	}
	return formatedSlots
}

// calculates slot placement by connverting
func calculateSlotPlacement(key string, records []record) int {
	low := 0
	high := len(records) - 1
	for low <= high {
		midPoint := low + (high-low)/2
		record := records[midPoint]
		if strings.Compare(key, record.key) < 0 {
			high = midPoint - 1
		} else {
			low = midPoint + 1
		}
	}
	return low
}

// Creates a buf containing the new and shifted slots
func getSlotReorderBuf(pageBytes []byte, slotBuf []byte, oldNumSlots int, slotPlacment int) []byte {
	numShiftedSlots := oldNumSlots - slotPlacment
	shiftedSlotsSize := numShiftedSlots * slotSize
	bufSize := shiftedSlotsSize + slotSize

	shiftedSlotsStartIndex := pageMetadataSize + (slotPlacment * slotSize)
	shiftedSlotsEndIndex := shiftedSlotsStartIndex + shiftedSlotsSize
	buf := make([]byte, bufSize)

	copy(buf, slotBuf)

	shiftSlotBytes := pageBytes[shiftedSlotsStartIndex:shiftedSlotsEndIndex]
	copy(buf[slotSize:], shiftSlotBytes)
	return buf
}

// Get offset for new and shift slot buf
func getSlotOffset(slotsBeforeNewslot int) int {
	slotsBeforeNewslotSize := slotsBeforeNewslot * slotSize
	slotOffset := slotsBeforeNewslotSize + pageMetadataSize
	return slotOffset
}
