package database

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func logAction(DB *DB, message string) error {
	_, err := DB.File.WriteString(message + "\n")
	if err != nil {
		return fmt.Errorf("Unable to log action")
	}
	return err
}

func replyLog(f *os.File) (map[string]string, error) {
	f.Seek(0, 0)
	scanner := bufio.NewScanner(f)
	data, err := processLogFile(scanner)
	if err := scanner.Err(); err != nil {
		return data, err
	}
	return data, err
}

func processLogFile(scanner *bufio.Scanner) (map[string]string, error) {
	// Note I am assume all key and value have no spaces
	// I are space it will break (I will fix this later on)
	// commard (SET DELETE), key, value
	var err error
	data := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		info := strings.Fields(line)

		if info[0] == "SET" && len(info) >= 3 {
			data[info[1]] = info[2]
		}
		if info[0] == "DELETE" && len(info) == 2 {
			delete(data, info[1])
		}
	}
	return data, err
}
