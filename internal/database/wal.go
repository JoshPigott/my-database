package database

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	uncommited byte = 0
	commited   byte = 1
	commitSize      = 1

	offsetSize   = 8
	bytesLenSize = 4

	offsetIndex   = 0
	bytesLenIndex = 8

	WALFilename
)

type WALFile struct {
	*os.File
}

func openWALFile(filename string) (WALFile, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return WALFile{}, fmt.Errorf("failed to open wal file: %w", err)
	}
	f := WALFile{file}
	info, err := file.Stat()
	if err != nil {
		fmt.Println("failed to get file size: %w", err)
	}
	// File was just made
	if info.Size() == 0 {
		f.setToUncommited()
	}
	return f, err
}

// Write a pages bytes into WAL file with page offset
func (f WALFile) writeEntry(offset int64, buf []byte) error {
	entryBufSize := offsetSize + len(buf)
	entryBuf := make([]byte, entryBufSize)

	binary.BigEndian.PutUint64(entryBuf[offsetIndex:offsetIndex+offsetSize], uint64(offset))
	copy(entryBuf[offsetSize:], buf)

	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("failed to seek end of WAL: %w", err)
	}
	if _, err := f.Write(entryBuf); err != nil {
		return fmt.Errorf("failed to write buf entry: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync wal file: %w", err)
	}
	return nil
}

// So marks all wal file to start commiting
func (pages Pages) commit() error {
	if _, err := pages.walFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek start of WAL: %w", err)
	}
	bytes := []byte{commited}
	if _, err := pages.walFile.Write(bytes); err != nil {
		return fmt.Errorf("failed to write commit to WAL: %w", err)
	}
	if err := pages.File.Sync(); err != nil {
		return fmt.Errorf("failed to sync wal file: %w", err)
	}
	pages.reply()
	return nil
}

// Writes the WAL file chagnes to main database
func (pages Pages) reply() error {
	offset := 0
	if _, err := pages.walFile.Seek(commitSize, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek start of WAL: %w", err)
	}
	content, err := io.ReadAll(pages.walFile)
	if err != nil {
		return fmt.Errorf("failed to read WAL file: %w", err)
	}

	for offset < len(content) {
		bytesOffset := int64(binary.BigEndian.Uint64(content[offset : offset+offsetSize]))
		offset += offsetSize
		bytes := content[offset : offset+pageSize]
		offset += pageSize

		if _, err := pages.File.Seek(bytesOffset, io.SeekStart); err != nil {
			return fmt.Errorf("failed to seek to offset: %w", err)
		}
		if _, err := pages.File.Write(bytes); err != nil {
			return fmt.Errorf("failed to write bytes: %w", err)
		}
		if err := pages.File.Sync(); err != nil {
			return fmt.Errorf("failed to sync wal file: %w", err)
		}
	}
	if err := pages.walFile.reset(); err != nil {
		return fmt.Errorf("failed to rest wal file: %w", err)
	}
	return nil
}

// If commited replys wal file else clears it
func (pages Pages) needsReply() error {
	info, err := os.Stat(pages.walFile.Name())
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	bytes := make([]byte, commitSize)
	if _, err := pages.walFile.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to offset: %w", err)
	}
	if _, err := pages.walFile.Read(bytes); err != nil {
		return fmt.Errorf("failed to read wal file: %w", err)
	}
	commitByte := bytes[0]
	if commitByte == uncommited && info.Size() == commitSize {
		return nil
	}
	if commitByte == uncommited {
		pages.walFile.reset()
	}
	return pages.reply()
}

// Reset the WAL to being uncommit with no writes
func (f WALFile) reset() error {
	os.Truncate(f.Name(), 0)
	f.setToUncommited()
	return nil
}

// Sets WAL to being uncommit with no writes
func (f WALFile) setToUncommited() error {
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek to offset: %w", err)
	}
	if _, err := f.Write([]byte{uncommited}); err != nil {
		return fmt.Errorf("failed to write bytes: %w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("failed to sync wal file: %w", err)
	}
	return nil
}
