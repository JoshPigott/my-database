package main

import (
	"bubbly-database/internal/database"
	"fmt"
)

func main() {
	DB, err := database.OpenDatabase()
	if err != nil {
		err = fmt.Errorf("failed to open the database: %w", err)
		panic(err)
	}
	for range 26 {
		err = DB.Set("a", "1")
		if err != nil {
			println("failed to add data", err)
		}
		err = DB.Set("a", "2")
		if err != nil {
			println("failed to add data", err)
		}
		err = DB.Set("b", "3")
		if err != nil {
			println("failed to add data", err)
		}
		err = DB.Delete("a")
		if err != nil {
			println("failed to delete data")
		}
	}
	fmt.Println("b:", DB.Get("b"))
	DB.Close()
}
