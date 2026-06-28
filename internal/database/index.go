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
	t, err := pages.getTree()
	if err != nil {
		return nil, fmt.Errorf("failed to get b+tree: %w", err)
	}
	db := &DB{
		Pages: pages,
		T:     t,
	}
	return db, nil
}

// Get tree or makes tree
func (pages *Pages) getTree() (*BTree, error) {
	var t *BTree
	rootID, isRoot, err := pages.getRootPage() // Bad here
	if err != nil {
		return nil, fmt.Errorf("failed to get root page: %w", err)
	}

	if isRoot == false {
		t, err = pages.NewBTree()
		if err != nil {
			return nil, err
		}
		if err := t.root.rootPageLink(); err != nil {
			return nil, fmt.Errorf("failed to create new b+tree: %w", err)
		}
	} else {
		rootNode, err := pages.ReadPageNode(rootID)
		if err != nil {
			return nil, fmt.Errorf("failed to read root node: %w", err)
		}
		t = &BTree{
			root: rootNode,
		}
	}
	return t, nil
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
