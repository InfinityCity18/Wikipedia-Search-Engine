package main

import (
	"fmt"
	"log/slog"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/database"
)

func main() {
	db, err := database.Create("postgres://hyperbarq:mownit@localhost:5432/search_engine", "/home/hyperbarq/Documents/wyszaszarka/PlainTextWikipedia/processed")
	if err != nil {
		slog.Error("Failed to create database", "error", err)
		return
	}
	art, err := db.ProcessQuery("final fantasy")
	if err != nil {
		fmt.Println(err)
	} else {
		for _, article := range art {
			fmt.Println(*article)
		}
	}
}
