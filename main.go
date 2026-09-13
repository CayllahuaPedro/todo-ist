package main

import (
	"fmt"
	"log"
	"todo-ist/database"
)

func main() {
	db, err := database.InitDB("todo-ist.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	fmt.Println("Database initialized successfully")

}
