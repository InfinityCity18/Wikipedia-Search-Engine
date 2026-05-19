package main

import (
	"log/slog"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/database"
	"github.com/InfinityCity18/Wikipedia-Search-Engine/webserver"
)

const documents_path = "/home/hyperbarq/Documents/wyszaszarka/PlainTextWikipedia/processed"
const postgres_path = "postgres://hyperbarq:mownit@localhost:5432/search_engine"

func main() {
	db, err := database.Create(postgres_path, documents_path)
	if err != nil {
		slog.Error("Failed to create database", "error", err)
		return
	}
	webserver.StartWebserver(&db)
}
