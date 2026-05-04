package database

import (
	"fmt"
	"os"
)

type DB struct {
	Data map[string]string
	File *os.File
}

func OpenDatabase() (*DB, error) {
	// This is wrong I will not able to append
	f, err := os.OpenFile("data/database-log.txt", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return nil, err
	}
	data, err := replyLog(f)

	DB := DB{
		Data: data,
		File: f,
	}
	return &DB, err
}

func (DB *DB) Close() error {
	err := DB.File.Close()
	return err
}

func (DB *DB) Set(key string, value string) error {
	// Updates log
	message := fmt.Sprintf("SET %s %s", key, value)
	err := logAction(DB, message)
	// Updates data
	DB.Data[key] = value
	return err
}

func (DB *DB) Update(key string, value string) error {
	// Updates log
	message := fmt.Sprintf("SET %s %s", key, value)
	err := logAction(DB, message)
	// Updates data
	DB.Data[key] = value
	return err
}

func (DB *DB) Delete(key string) error {
	// Updates log
	message := fmt.Sprintf("DELETE %s", key)
	err := logAction(DB, message)
	// Updates data
	delete(DB.Data, key)
	return err
}

// Convert map into a list
func (DB *DB) GetAll() [][]string {
	allData := [][]string{}
	for key, value := range DB.Data {
		item := []string{key, value}
		allData = append(allData, item)
	}
	return allData
}

func (DB *DB) Get(key string) string {
	return DB.Data[key]
}
