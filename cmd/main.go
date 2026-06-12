package main

import (
	"bubbly-database/internal/btrees"
	"bubbly-database/internal/database"
	"fmt"
	"io"
	"strconv"
)

var DB *database.DB

func main() {
	var entries = []struct {
		key    string
		pageID uint32
		slotID uint32
	}{
		{key: "X7mK", pageID: 2, slotID: 14},
		{key: "Q9pL", pageID: 5, slotID: 3},
		{key: "B4vN", pageID: 8, slotID: 21},
		{key: "T8rJ", pageID: 11, slotID: 7},
		{key: "M2xF", pageID: 1, slotID: 18},
		{key: "K6wP", pageID: 6, slotID: 2},
		{key: "H3nR", pageID: 9, slotID: 25},
		{key: "V7qD", pageID: 3, slotID: 11},
		{key: "L5zC", pageID: 7, slotID: 30},
		{key: "N8bY", pageID: 12, slotID: 6},
		{key: "P4tG", pageID: 4, slotID: 16},
		{key: "R9kW", pageID: 10, slotID: 22},
		{key: "X7mK9pL2", pageID: 2, slotID: 14},
		{key: "Q9pL4vNb", pageID: 5, slotID: 3},
		{key: "B4vN8rJ6", pageID: 8, slotID: 21},
		{key: "T8rJ2xFm", pageID: 11, slotID: 7},
		{key: "M2xF7qDh", pageID: 1, slotID: 18},
		{key: "K6wP3nRv", pageID: 6, slotID: 2},
		{key: "H3nR5zCy", pageID: 9, slotID: 25},
		{key: "V7qD8bYk", pageID: 3, slotID: 11},
		{key: "L5zC4tGp", pageID: 7, slotID: 30},
		{key: "N8bY9kWr", pageID: 12, slotID: 6},
		{key: "P4tG6mXs", pageID: 4, slotID: 16},
		{key: "R9kW2pLv", pageID: 10, slotID: 22},
	}

	t := btrees.NewBTree()
	for _, e := range entries {
		loc := btrees.KeyLocation{PageID: e.pageID, SlotID: e.slotID}
		t.Insert(e.key, &loc)
	}
	t.CheckStructure(1)
}

func testDatabase() {
	var err error
	DB, err = database.Open()
	if err != nil {
		fmt.Println(err)
	}
	for i := 5; i < 9; i++ {
		theString := strconv.Itoa(i)
		theValue := theString + theString
		if err := DB.AddToPage(theString, theValue); err != nil {
			fmt.Println("failed to add to page:", err)
		}
	}
	DB.AddToPage("a", "67")
	// DB.Delete("1")
	// lookAtFile()
	lookAtData()
	if err := DB.Close(); err != nil {
		fmt.Println(err)
	}
}

func lookAtFile() {
	info, err := DB.File.Stat()
	if err != nil {
		fmt.Println("failed to get file size: %w", err)
	}
	fileSize := info.Size()
	fmt.Println("info.Size()", fileSize)
	// for key, value := range data {
	// 	fmt.Println("key:", key, "value:", value)
	// }
	bytes := make([]byte, fileSize)
	if _, err := DB.File.Seek(0, io.SeekStart); err != nil {
		fmt.Println("failed to move start of file to read")
	}
	n, err := DB.File.Read(bytes)
	if err != nil {
		fmt.Println("failed to read bytes aye", err)
	}

	// I need to seek to the start here

	fmt.Println("Number of bytes read:", n)
	for i := 0; i < int(fileSize)/4096; i++ {
		fmt.Println()
		fmt.Println("page:", (i + 1))
		fmt.Println(bytes[i*4096 : (i+1)*4096])
	}
	// fmt.Println(bytes)

}

func lookAtData() {
	data, err := DB.SelectAll()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Map length:", len(data))
	for key, value := range data {
		fmt.Println("key:", key, "value:", value)
	}
}
