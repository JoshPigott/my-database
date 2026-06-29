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

type DB struct {
	Pages *Pages
	Root  *node
}

// Opens up database file a makes sure there is a metedata page
func openDefault(filename string) (*DB, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open database file: %w", err)
	}
	db, err := newDatabase(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create new database: %w", err)
	}
	return db, nil
}

func newDatabase(file *os.File) (*DB, error) {
	pages := &Pages{
		File:         file,
		pageIDToNode: map[uint32]*node{},
	}

	if err := pages.ensureDBMetadataPage(file); err != nil {
		return nil, fmt.Errorf("failed to create metadata page: %w", err)
	}
	if err := pages.createFirstDataPage(); err != nil {
		return nil, fmt.Errorf("failed to create first data page: %w", err)
	}
	root, err := pages.getRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to get b+tree root: %w", err)
	}
	db := &DB{
		Pages: pages,
		Root:  root,
	}
	return db, nil
}

// Get tree root node from disk or make it
func (pages *Pages) getRoot() (*node, error) {
	var root *node
	rootID, isRoot, err := pages.getRootPage()
	if err != nil {
		return nil, fmt.Errorf("failed to get root page: %w", err)
	}

	if !isRoot {
		root, err = pages.newNode(true)
		if err != nil {
			return nil, fmt.Errorf("failed to create new root node: %w", err)
		}
		if err := root.rootPageLink(); err != nil {
			return nil, fmt.Errorf("failed to add root id in database: %w", err)
		}
	} else {
		root, err = pages.ReadPageNode(rootID)
		if err != nil {
			return nil, fmt.Errorf("failed to read root node: %w", err)
		}
	}
	return root, nil
}
