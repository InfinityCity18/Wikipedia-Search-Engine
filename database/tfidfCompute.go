package database

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/james-bowman/sparse"
)

type Item struct {
	Id   int
	Word string
	Idf  float64
}

type MatrixWrite struct {
	Row int
	Col int
	Val float64
}

func (db *Database) computeTfidf(art_list []*articles.Article) error {
	rows, err := db.pool.Query(context.Background(), `
		SELECT id, word, idf FROM dictionary
	`)
	if err != nil {
		slog.Error("Failed to get words from table", "error", err)
	}
	index_lookup := make(map[string]int)
	words := []Item{}

	for rows.Next() {
		var item Item

		err := rows.Scan(&item.Id, &item.Word, &item.Idf)
		if err != nil {
			slog.Error("Failed to scan", "error", err)
		}
		index_lookup[item.Word] = item.Id - 1 //sql db uses 1 indexing
		words = append(words, item)
	}
	amount_of_words := len(words)
	matrix := sparse.NewDOK(amount_of_words, len(art_list))
	var wg sync.WaitGroup
	ch := make(chan MatrixWrite)

	for i, article := range art_list {
		wg.Add(1)
		go func(i int, article *articles.Article) {
			defer wg.Done()
			words_map := make(map[string]int)
			total := float64(len(strings.Fields(article.Stemmed)))
			for word := range strings.FieldsSeq(article.Stemmed) {
				_, ok := words_map[word]
				if !ok {
					words_map[word] = 1
				} else {
					words_map[word]++
				}
			}
			for word, count := range words_map {
				data := MatrixWrite{Row: index_lookup[word], Col: i, Val: float64(count) / total * words[index_lookup[word]].Idf}
				ch <- data
			}
		}(i, article)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	for data := range ch {
		matrix.Set(data.Row, data.Col, data.Val)
	}
	fmt.Println(matrix)
	return nil
}
