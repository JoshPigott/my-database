package database

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const temporaryLogFile = "data/temp-database-log.txt"

// Logs action to a database log
func (DB *DB) logAction(message string) error {
	_, err := DB.File.WriteString(message + "\n")
	if err != nil {
		return fmt.Errorf("failed to log action: %w", err)
	}
	// Makes sure operation get rewriten to disk even if code fails
	DB.File.Sync()
	err = DB.updateOperationsNum()
	return err
}

// Increase operation count by one and checks if log compaction is need
func (DB *DB) updateOperationsNum() error {
	if DB.Metadata.OperationsSinceLastCompaction >= 100 {
		err := DB.compactLog()
		if err != nil {
			return fmt.Errorf("failed to compact log file: %w", err)
		}
		DB.Metadata.OperationsSinceLastCompaction = 0
	} else {
		DB.Metadata.OperationsSinceLastCompaction += 1
	}
	if err := DB.Metadata.Update(); err != nil {
		return fmt.Errorf("failed to update metdata: %w", err)
	}
	return nil
}

// Reply logs to get data
func replyLog(f *os.File) (map[string]string, error) {
	f.Seek(0, 0)
	scanner := bufio.NewScanner(f)
	data := processLogFile(scanner)
	if err := scanner.Err(); err != nil {
		return data, fmt.Errorf("failed to read log %w", err)
	}
	return data, nil
}

// Rewrites log file with a temp file removing unneed lines
func (DB *DB) compactLog() error {
	if err := DB.tempLogFile(); err != nil {
		return err
	}
	if err := DB.switchLogFile(); err != nil {
		return err
	}
	return nil
}

// Create temp log and writes data to it
func (DB *DB) tempLogFile() error {
	tempFile, err := os.Create(temporaryLogFile)
	if err != nil {
		return fmt.Errorf("failed to create temporary log file: %w", err)
	}
	for key, value := range DB.Data {
		if _, err = fmt.Fprintf(tempFile, "SET %s %s\n", key, value); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to write temporary log file: %w", err)
		}
	}
	if err = tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write temporary log file: %w", err)
	}
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temporary log file: %w", err)
	}
	return nil
}

// Close log file, then swtich with temporary file then reopens
func (DB *DB) switchLogFile() error {
	if err := DB.File.Close(); err != nil {
		return fmt.Errorf("failed to close log file for switching: %w", err)
	}
	if err := os.Rename(temporaryLogFile, logFileName); err != nil {
		return fmt.Errorf("failed to switch log file with temporary log file: %w", err)
	}
	f, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0664)
	if err != nil {
		return fmt.Errorf("failed to open log file after compaction: %w", err)
	}
	DB.File = f
	return nil
}

// Process each line of log file rebuilding the data
func processLogFile(scanner *bufio.Scanner) map[string]string {
	data := map[string]string{}
	for scanner.Scan() {
		line := scanner.Text()
		info := strings.Fields(line)

		if len(info) >= 3 && info[0] == "SET" {
			data[info[1]] = info[2]
		}
		if len(info) == 2 && info[0] == "DELETE" {
			delete(data, info[1])
		}
	}
	return data
}
