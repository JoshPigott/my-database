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

	if err := pages.ensureMetadataPage(); err != nil {
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
