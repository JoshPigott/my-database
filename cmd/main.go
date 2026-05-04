package main

import (
	"bubbly-database/internal/database"
	"fmt"
)

func main() {
	DB, err := database.OpenDatabase()
	if err != nil {
		panic("Unable to open the database")
	}
	err = DB.Set("a", "1")
	if err != nil {
		println("Unable to add data")
	}
	err = DB.Set("a", "2")
	if err != nil {
		println("Unable to add data")
	}
	fmt.Println("a:", DB.Get("a"))
	fmt.Println("All data", DB.GetAll())
	err = DB.Delete("a")
	if err != nil {
		println("Unable to delete data")
	}
	DB.Close()
}

// database.InsertData("Josh", "Just try")
// database.InsertData("Josh", "Josh?...")
// database.InsertData("Chicken", "Dead")

// database.SelectData()
