package main

import (
	"bubbly-database/internal/database"
	"fmt"
)

var DB *database.DB

func main() {
	DB, err := database.Open()
	if err != nil {
		fmt.Println(err)
	}
	if err = DB.CreatePage(database.MetadataPage); err != nil {
		fmt.Println(err)
	}
	// if err := DB.AddToPage("1", "one"); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := DB.AddToPage("12", "onetwo"); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := DB.AddToPage("1234", "4321"); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := DB.AddToPage("6789", "9876"); err != nil {
	// 	fmt.Println(err)
	// }
	// if err := DB.Delete("12"); err != nil {
	// 	fmt.Println(err)
	// }
	data, err := DB.SelectAll()
	if err != nil {
		fmt.Println(err)
	}
	for key, value := range data {
		fmt.Println("key:", key, "value:", value)
	}
	if err := DB.Close(); err != nil {
		fmt.Println(err)
	}
}
