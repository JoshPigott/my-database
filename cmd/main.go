package main

import (
	"bubbly-database/internal/pages"
	"fmt"
)

func main() {
	if err := pages.CreatePage(pages.MetadataPage); err != nil {
		fmt.Println("Create:", err)
	}
	for i := range 130 {
		if err := pages.AddToPage("Josh Pigott", "Can do anything"); err != nil {
			fmt.Println("Add:", err)
			fmt.Println("i:", i)
			break
		}
	}
	data, err := pages.ReadPage()
	if err != nil {
		fmt.Println("Read:", err)
	}
	for key, value := range data {
		println("Key:", key, "Value:", value)
	}
}
