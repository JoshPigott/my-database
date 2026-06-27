package database

import (
	"fmt"
	"os"
)

const pageStart = 0

type Pages struct {
	File         *os.File
	pageIDToNode map[uint32]*node
}

type Btree struct {
	T *BTree
}

type DB struct {
	Pages *Pages
	T     *BTree
}

func newDatabase(file *os.File) (*DB, error) {
	var t *BTree
	pages := &Pages{
		File:         file,
		pageIDToNode: map[uint32]*node{},
	}

	if err := pages.ensureMetadataPage(file); err != nil {
		return nil, fmt.Errorf("failed to create metadata page: %w", err)
	}
	if err := pages.createFirstDataPage(); err != nil {
		return nil, fmt.Errorf("failed to create first data page")
	}

	rootID, isRoot, err := pages.getRootPage()
	if err != nil {
		return nil, err
	}
	if isRoot == false {
		t, rootID, err = pages.NewBTree()
		if err != nil {
			return nil, err
		}
	} else {
		rootNode, err := pages.ReadPageNode(rootID)
		if err != nil {
			return nil, fmt.Errorf("failed to read node: %w", err)
		}
		t = &BTree{
			root: rootNode,
		}
	}

	if err := pages.rootPageLink(rootID); err != nil {
		return nil, err
	}

	db := &DB{
		Pages: pages,
		T:     t,
	}
	return db, nil
}

// Opens up database file a makes sure there is a metedata page
func openDefault(filename string) (*DB, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open page: %w", err)
	}
	db, err := newDatabase(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	return db, nil
}
