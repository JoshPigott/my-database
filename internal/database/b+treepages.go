package database

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Turns node page into free page
func (pages *Pages) deleteNodePage(n *node) error {
	buf := make([]byte, pageSize)

	metadata, err := pages.ReadMetadata()
	if err != nil {
		return fmt.Errorf("failed to get metadata: %w", err)
	}
	binary.BigEndian.PutUint32(buf[pageStart:pageIDSize], metadata.nextFreePage)
	metadata.nextFreePage = n.pageID

	if err := pages.updateMetadata(metadata); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	if err := pages.WriteBytes(buf, pageStart, n.pageID); err != nil {
		return fmt.Errorf("failed to delete nodes page: %w", err)
	}
	return nil
}

func (pages *Pages) ReadPageNode(pageID uint32) (*node, error) {
	var n *node
	offset := 0
	pageBytes, err := pages.ReadBytes(pageSize, offset, pageID)
	// Free page node was deleted
	if pageBytes[pageStart] == 0 {
		return nil, errors.New("invalid page node")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to node's page: %w", err)
	}
	metadataBytes := pageBytes[pageStart:pageMetadataSize]
	pageMetadata := formatPageMetadata(metadataBytes)
	switch pageMetadata.pageType {
	case InternalPage:
		n = pages.formatInternalNode(pageBytes, pageMetadata)
	case LeafPage:
		n = pages.formatLeafNode(pageBytes, pageMetadata)
	}
	n.pageID = pageID
	pages.pageIDToNode[pageID] = n
	return n, nil
}

// Loops over number of keys format getting the keys and children
func (pages *Pages) formatInternalNode(pageBytes []byte, pageMetadata pageMetadata) *node {
	n := node{}
	n.pages = pages
	offset := pageMetadataSize
	// Loops over number of keys
	for range pageMetadata.numEntries {
		offset = pages.formatInternalKey(&n, pageBytes, offset)
	}
	// Get last child. As one more children than keys
	if pageMetadata.numEntries > 0 {
		pages.formatChild(&n, pageBytes, offset)
	}
	return &n
}

func (pages *Pages) formatInternalKey(n *node, pageBytes []byte, offset int) int {
	pages.formatChild(n, pageBytes, offset)
	offset += pageIDSize

	// Get key length
	keylen := int(binary.BigEndian.Uint16(pageBytes[offset : offset+keyLenStorageSize]))
	offset += keyLenStorageSize

	keyBytes := pageBytes[offset : offset+keylen]
	offset += keylen
	key := string(keyBytes)
	n.keys = append(n.keys, key)
	return offset
}

func (pages *Pages) formatLeafNode(pageBytes []byte, pageMetadata pageMetadata) *node {
	n := node{}
	n.leaf = true
	offset := pageMetadataSize
	n.NextID = binary.BigEndian.Uint32(pageBytes[offset : offset+pageIDSize])
	offset += pageIDSize
	n.Next = pages.pageIDToNode[n.NextID]
	n.pages = pages

	for range pageMetadata.numEntries {
		offset = formatLeafKey(&n, pageBytes, offset)
	}
	return &n
}

// Format a leaf key
func formatLeafKey(n *node, pageBytes []byte, offset int) int {
	// Get key length
	keylen := int(binary.BigEndian.Uint16(pageBytes[offset : offset+keyLenStorageSize]))
	offset += keyLenStorageSize

	// Get key
	keyBytes := pageBytes[offset : offset+keylen]
	offset += keylen
	key := string(keyBytes)
	n.keys = append(n.keys, key)

	// Get key location
	pageID := binary.BigEndian.Uint32(pageBytes[offset : offset+pageIDSize])
	offset += pageIDSize
	slotID := binary.BigEndian.Uint16(pageBytes[offset : offset+slotIDSize])
	offset += slotIDSize
	n.keyLocations = append(n.keyLocations, newKeyLocation(pageID, slotID))
	return offset
}

func (pages *Pages) formatChild(n *node, pageBytes []byte, offest int) {
	childPageID := binary.BigEndian.Uint32(pageBytes[offest : offest+pageIDSize])
	n.childPageIDs = append(n.childPageIDs, childPageID)
	n.children = append(n.children, pages.pageIDToNode[childPageID])
}

func (pages *Pages) writeNodeToPage(n *node) {
	delete(pages.pageIDToNode, n.pageID)
	if n.leaf == true {
		pages.writeLeafNode(n)
	} else {
		pages.writeInternalNode(n)
	}
}

// Creates a whole leaf node to disk
func (Pages *Pages) writeLeafNode(n *node) {
	buf := make([]byte, pageSize)

	offset := pageMetadataSize
	binary.BigEndian.PutUint32(buf[offset:offset+pageIDSize], n.NextID)
	offset += pageIDSize

	for i, key := range n.keys {
		keyLocation := *n.keyLocations[i]
		offset = addLeafKey(&buf, keyLocation.PageID, keyLocation.SlotID, offset, key)
	}

	pageMetadata := pageMetadata{
		pageType:       LeafPage,
		pageID:         n.pageID,
		numEntries:     uint16(len(n.keys)),
		freeSpaceStart: uint16(offset),
		freeSpaceEnd:   pageSize,
	}
	metadataBuf := createMetadataBuffer(pageMetadata)
	copy(buf[pageStart:pageMetadataSize], metadataBuf)
	Pages.WriteBytes(buf, pageStart, n.pageID)
}

// Creates a whole internal node to disk
func (Pages *Pages) writeInternalNode(n *node) {
	buf := make([]byte, pageSize)
	offset := pageMetadataSize

	for i, key := range n.keys {
		offset = addInternalKey(&buf, offset, key, n.childPageIDs[i])
	}
	// Number of children is one great then number of keys
	binary.BigEndian.PutUint32(buf[offset:offset+pageIDSize], n.childPageIDs[len(n.childPageIDs)-1])
	offset += pageIDSize

	pageMetadata := pageMetadata{
		pageType:       InternalPage,
		pageID:         n.pageID,
		numEntries:     uint16(len(n.keys)),
		freeSpaceStart: uint16(offset),
		freeSpaceEnd:   pageSize,
	}
	metadataBuf := createMetadataBuffer(pageMetadata)
	copy(buf[pageStart:pageMetadataSize], metadataBuf)
	Pages.WriteBytes(buf, pageStart, n.pageID)
}

func addLeafKey(buf *[]byte, pageID uint32, slotID uint16, offset int, key string) int {
	// Key len
	binary.BigEndian.PutUint16((*buf)[offset:offset+keyLenStorageSize], uint16(len(key)))
	offset += keyLenStorageSize

	// Key
	keyBytes := []byte(key)
	copy((*buf)[offset:offset+len(key)], keyBytes)
	offset += len(key)

	// PageID
	binary.BigEndian.PutUint32((*buf)[offset:offset+pageIDSize], pageID)
	offset += pageIDSize

	// SlotID
	binary.BigEndian.PutUint16((*buf)[offset:offset+slotIDSize], slotID)
	offset += slotIDSize
	return offset
}

func addInternalKey(buf *[]byte, offset int, key string, childPageID uint32) int {
	// Child node page
	binary.BigEndian.PutUint32((*buf)[offset:offset+pageIDSize], childPageID)
	offset += pageIDSize

	// Key len
	binary.BigEndian.PutUint16((*buf)[offset:offset+keyLenStorageSize], uint16(len(key)))
	offset += keyLenStorageSize

	// Key
	keyBytes := []byte(key)
	copy((*buf)[offset:offset+len(key)], keyBytes)
	offset += len(key)
	return offset
}

// Checks if data is less than half a page size
func (n *node) isUnderflow() bool {
	currSize := n.getDataSize()
	return currSize < halfPageSize
}

// Checks if all the data will fit on one page or not
func (n *node) isOverflow() bool {
	currSize := n.getDataSize()
	return currSize > pageSize
}

func (n *node) getDataSize() int {
	currSize := pageMetadataSize
	if n.leaf {
		currSize += pageIDSize
		for _, key := range n.keys {
			currSize += keyLenStorageSize
			currSize += len(key)
			currSize += slotIDSize
			currSize += pageIDSize
		}
	} else {
		for _, key := range n.keys {
			currSize += keyLenStorageSize
			currSize += len(key)
			currSize += pageIDSize
		}
		currSize += pageIDSize
	}
	return currSize
}

// Check if child + left exceeds page size
func (n *node) needsLeftRedistribution(i int) bool {
	childSize := n.children[i].getDataSize()
	leftChildSize := n.children[i-1].getDataSize()
	totalSize := childSize + leftChildSize - pageMetadataSize
	return totalSize > pageSize
}

// Check if child + right exceeds page size
func (n *node) needsRightRedistribution(i int) bool {
	childSize := n.children[i].getDataSize()
	leftChildSize := n.children[i+1].getDataSize()
	totalSize := childSize + leftChildSize - pageMetadataSize
	return totalSize > pageSize
}
