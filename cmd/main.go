package main

import (
	"bubbly-database/internal/database"
	"fmt"
	"io"
	"strconv"
)

var DB *database.DB
var insertedKeys []string

func main() {
	testDatabase()
}

func testDatabase() {
	var err error
	DB, err = database.Open()
	if err != nil {
		fmt.Println(err)
	}
	num := 1000000000000000000
	amount := 500
	half := strconv.Itoa(num + amount/2)
	for i := 0; i < amount; i += 1 {
		key := strconv.Itoa(i + num)
		val := strconv.Itoa(i)
		DB.AddToPage(key, val)
	}
	DB.Delete("1000000000000000132")
	DB.Delete(half)
	// allData, _ := DB.SelectAll()
	// fmt.Println("allData:", allData)
	// DB.T.PrintLinkedList()

	selectedData, err := DB.SelectWhere(database.GreaterThan, half)
	if err != nil {
		fmt.Println("failed to select data: %w", err)
	}
	fmt.Println("The selected data:", selectedData)
	if err := DB.Close(); err != nil {
		fmt.Println(err)
	}
}

func lookAtFile() {
	info, err := DB.Pages.File.Stat()
	if err != nil {
		fmt.Println("failed to get file size: %w", err)
	}
	fileSize := info.Size()
	fmt.Println("info.Size()", fileSize)
	// for key, value := range data {
	// 	fmt.Println("key:", key, "value:", value)
	// }
	bytes := make([]byte, fileSize)
	if _, err := DB.Pages.File.Seek(0, io.SeekStart); err != nil {
		fmt.Println("failed to move start of file to read")
	}
	n, err := DB.Pages.File.Read(bytes)
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
		fmt.Printf("key: %-12s value: %-22s\n", key, value)
	}
}
