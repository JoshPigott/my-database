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
		key    int
		pageID uint32
		slotID uint32
	}{
		{key: 7, pageID: 2, slotID: 14},
		{key: 13, pageID: 5, slotID: 3},
		{key: 19, pageID: 8, slotID: 21},
		{key: 3, pageID: 11, slotID: 7},
		{key: 25, pageID: 1, slotID: 18},
		{key: 30, pageID: 6, slotID: 2},
		{key: 11, pageID: 9, slotID: 25},
		{key: 22, pageID: 3, slotID: 11},
		{key: 17, pageID: 7, slotID: 30},
		{key: 5, pageID: 12, slotID: 6},
		{key: 28, pageID: 4, slotID: 16},
		{key: 9, pageID: 10, slotID: 22},
	}

	t := btrees.NewBTree()
	for _, e := range entries {
		loc := btrees.KeyLocation{PageID: e.pageID, SlotID: e.slotID}
		t.Insert(e.key, &loc)
	}

	t.Delete(13)
	t.Delete(17)
	t.Delete(11)
	t.Delete(3)
	t.Delete(28)
	t.Delete(30)
	t.Delete(9)
	t.Delete(5)
	t.Delete(7)
	t.Delete(19)
	t.Delete(25)
	t.Delete(22)

	for _, e := range entries {
		key := e.key
		keyLocation, found := t.FindKeyLocation(key)
		fmt.Println("Key:", key)
		fmt.Println("found:", found)
		fmt.Println("keyLocation.PageID:", keyLocation.PageID)
		fmt.Println("keyLocation.SlotID:", keyLocation.SlotID)
	}
	// key := 7
	// fmt.Println(key)
	// keyLocation, found := t.FindKeyLocation(key)
	// fmt.Println("Key:", key)
	// fmt.Println("found:", found)
	// fmt.Println("keyLocation.PageID:", keyLocation.PageID)
	// fmt.Println("keyLocation.SlotID:", keyLocation.SlotID)

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
