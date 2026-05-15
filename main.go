package main

import (
	"fmt"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/database"
)

func main() {
	_, err := database.Create("postgres://hyperbarq:mownit@localhost:5432/search_engine", "/home/hyperbarq/Documents/wyszaszarka/PlainTextWikipedia/processed")
	fmt.Println(err)
	v := 2 + 2
	_ = v
}
