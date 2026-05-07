package database

import (
	"fmt"
	"os"
)

const logFileName = "data/database-log.txt"

type DB struct {
	Data     map[string]string
	File     *os.File
	Metadata *Metadata
}

// So every time I change the meta do I need to rewrite the whole meta data file?
func OpenDatabase() (*DB, error) {
	// This is wrong I will not able to append
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return nil, err
	}

	metadata, err := getMetadata()
	if err != nil {
		return nil, err
	}

	data, err := replyLog(f)
	if err != nil {
		return nil, err
	}
	DB := DB{
		Data:     data,
		File:     f,
		Metadata: metadata,
	}
	return &DB, err
}

func (DB *DB) Close() error {
	err := DB.File.Close()
	// close meta file
	return err
}

func (DB *DB) Set(key string, value string) error {
	// Updates log
	message := fmt.Sprintf("SET %s %s", key, value)
	// Updates data
	DB.Data[key] = value
	err := DB.logAction(message)
	return err
}

func (DB *DB) Update(key string, value string) error {
	// Updates log
	message := fmt.Sprintf("SET %s %s", key, value)
	// Updates data
	DB.Data[key] = value
	err := DB.logAction(message)
	return err
}

func (DB *DB) Delete(key string) error {
	// Updates log
	message := fmt.Sprintf("DELETE %s", key)
	// Updates data
	delete(DB.Data, key)
	err := DB.logAction(message)
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
