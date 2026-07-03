package database

import (
	"errors"
	"fmt"
	"io"
)

var ErrNotDataPage = errors.New("failed to read page due to not data page")

// Read pages and get bytes from page
func (pages *Pages) ReadBytes(size int, offset int, pageID uint32) ([]byte, error) {
	pageBytes := pages.checkCache(pageID)
	if pageBytes != nil {
		return pageBytes[offset : offset+size], nil
	}

	pageBytes = make([]byte, pageSize)
	pageOffset := getPageOffset(pageID)
	if _, err := pages.File.Seek(pageOffset, io.SeekStart); err != nil {
		return []byte{}, fmt.Errorf("failed to seek to offset: %w", err)
	}
	if _, err := pages.File.Read(pageBytes); err != nil {
		return []byte{}, fmt.Errorf("failed to read bytes: %w", err)
	}
	pages.updateCache(pageID, pageBytes)
	return pageBytes[offset : offset+size], nil
}

// Open up database writes bytes and syncs bytes
func (pages *Pages) writeBytes(bytes []byte, offset int, pageID uint32) error {
	var pageBytes []byte
	if len(bytes)+offset > pageSize {
		return errors.New("more one than one page")
	}
	pageOffset := getPageOffset(pageID)
	if len(bytes) == pageSize {
		pageBytes = bytes
	} else {
		pageBytes = pages.checkCache(pageID)
		if pageBytes == nil {
			pageBytes = make([]byte, pageSize)
			if _, err := pages.File.Seek(pageOffset, io.SeekStart); err != nil {
				return fmt.Errorf("failed to seek to offset: %w", err)
			}
			if _, err := pages.File.Read(pageBytes); err != nil {
				return fmt.Errorf("failed to read page bytes: %w", err)
			}
		}
		copy(pageBytes[offset:offset+len(bytes)], bytes)
	}
	pages.walFile.writeEntry(pageOffset, pageBytes)
	pages.updateCache(pageID, pageBytes)
	return nil
}

// Check cache to get page bytes from memory
func (pages *Pages) checkCache(pageID uint32) []byte {
	page, found := pages.cache.items[pageID]
	if !found {
		return nil
	}
	pages.moveToFront(page)
	return page.bytes
}

// Adds or update pages cache; bytes and linked list
func (pages *Pages) updateCache(pageID uint32, pageBytes []byte) {
	page, found := pages.cache.items[pageID]
	if found {
		pages.moveToFront(page)
		return
	}
	// Update linked list
	oldHead := pages.cache.head
	page = &Page{
		id:    pageID,
		bytes: pageBytes,
		prev:  oldHead,
	}
	if oldHead == nil {
		pages.cache.tail = page
	} else {
		oldHead.next = page
	}
	pages.cache.head = page

	// Update page map
	pages.cache.items[pageID] = page

	// Free memory
	if len(pages.cache.items) > pages.cache.capacity {
		oldTail := pages.cache.tail
		newTail := oldTail.next
		oldTail.next = nil
		pages.cache.tail = newTail
		if newTail != nil {
			newTail.prev = nil
		} else {
			pages.cache.head = nil
		}

		delete(pages.cache.items, oldTail.id)
	}
}

// Move page up to the front now being last used page
func (pages *Pages) moveToFront(page *Page) {
	if page == pages.cache.head {
		return
	}
	// Update linked list
	if pages.cache.tail == page {
		pages.cache.tail = page.next
	}

	if page.next != nil {
		page.next.prev = page.prev
	}
	if page.prev != nil {
		page.prev.next = page.next
	}

	oldHead := pages.cache.head
	page.prev = oldHead
	page.next = nil
	oldHead.next = page
	pages.cache.head = page
}

// Reads page metadata (I will need to put what page later on)
func (pages *Pages) readPageMetadata(pageID uint32) (pageMetadata, error) {
	metadataBytes, err := pages.ReadBytes(pageMetadataSize, pageStart, pageID)
	if err != nil {
		return pageMetadata{}, fmt.Errorf("failed to read page metadata: %w", err)
	}
	pageMetadata := formatPageMetadata(metadataBytes)
	return pageMetadata, nil
}

// Read page to get the data in a data page
func (pages *Pages) readDataPage(pageID uint32) ([]record, error) {
	pageBytes, err := pages.ReadBytes(pageSize, pageStart, pageID)
	if err != nil {
		return []record{}, fmt.Errorf("failed to read bytes: %w", err)
	}

	// format data
	pageMetadata := formatPageMetadata(pageBytes)
	if pageMetadata.pageType != DataPage {
		return []record{}, ErrNotDataPage
	}
	slots := formatSlots(pageBytes, pageMetadata)
	dataRecords := formatData(pageBytes, slots)
	return dataRecords, nil
}

func getPageOffset(pageID uint32) int64 {
	offset := (int64(pageID) - 1) * pageSize
	return offset
}
