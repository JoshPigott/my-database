package database

import (
	"fmt"
	"os"
)

// THere are going to be querrys that the user can send

type DB struct {
	File *os.File
}

func newDatabase(file *os.File) DB {
	return DB{
		File: file,
	}
}

func Open() (*DB, error) {
	file, err := os.OpenFile(FileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open page: %w", err)
	}
	db := newDatabase(file)
	return &db, nil

}

func (DB *DB) Close() error {
	if err := DB.File.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}
	return nil
}
