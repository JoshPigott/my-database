package main

import (
	"bubbly-database/internal/database"
	"fmt"
	"io"
	"strconv"
)

var DB *database.DB

func main() {
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
