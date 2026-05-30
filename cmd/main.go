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
	for i := 0; i < 400; i++ {
		theString := strconv.Itoa(i)
		theValue := theString + theString
		if err := DB.AddToPage(theString, theValue); err != nil {
			fmt.Println("failed to add to page:", err)
		}
	}
	DB.Delete("270")
	DB.WriteSix()
	lookAtFile()
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
	data, err := DB.SelectAll()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Map length:", len(data))
	for key, value := range data {
		fmt.Println("key:", key, "value:", value)
	}
}

// Test small way untill it breaks
// Like adding more slowy

// if err := DB.AddToPage("6789", "9876"); err != nil {
// 	fmt.Println(err)
// }
// if err := DB.Delete("12"); err != nil {
// 	fmt.Println(err)
// }
// data, err := DB.SelectAll()
// if err != nil {
// 	fmt.Println(err)
// }
// for key, value := range data {
// 	fmt.Println("key:", key, "value:", value)
// }

// if err := DB.AddToPage("6.7", "onetwo"); err != nil {
// 	fmt.Println(err)
// }
