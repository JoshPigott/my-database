package main

import (
	"bubbly-database/internal/database"
	"fmt"
	"io"
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
	addKeysValues()
	lookAtFile()

	val, _, _ := DB.Select("dataset_worksheet_user_profile_system_generated_long_form_key_field_identifier__0500_v1")
	fmt.Println("value", val)
	// lookAtFile()
	// lookAtData()
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

func deleteKeysValues() {
	for _, key := range insertedKeys {
		// fmt.Printf("DB.Delete(%s)\n", key)
		DB.Delete(key)
	}
}

// For testing: add lots of different key values
func addKeysValues() {
	insertedKeys = insertedKeys[:0] // Clear any previous keys

	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf(
			"dataset_worksheet_user_profile_system_generated_long_form_key_field_identifier__%04d_v1",
			i,
		)
		val := fmt.Sprintf("%02x", i%256)

		// fmt.Printf("DB.AddToPage(\"%s\", \"%s\")\n", key, val)
		DB.AddToPage(key, val)
		insertedKeys = append(insertedKeys, key)
	}
}
