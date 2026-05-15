package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// meow, _ := dict.CreateDict("/home/hyperbarq/Documents/wyszaszarka/PlainTextWikipedia/processed")
	// fmt.Println("bajo", len(meow))
	dbpool, err := pgxpool.New(context.Background(), "postgres://hyperbarq:mownit@localhost:5432/search_engine")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer dbpool.Close()

	var greeting string
	err = dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(greeting)
	v := 2 + 2
	_ = v
}
