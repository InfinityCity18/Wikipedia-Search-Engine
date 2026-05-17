package database

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"sync"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/InfinityCity18/Wikipedia-Search-Engine/redsvd"
	"github.com/pgvector/pgvector-go"
	"gonum.org/v1/gonum/mat"
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

func (db *Database) insertDocuments(art_list []*articles.Article) (*mat.Dense, error) {
	rows, err := db.pool.Query(context.Background(), `
		SELECT id, word, idf FROM dictionary
	`)
	if err != nil {
		slog.Error("Failed to get words from table", "error", err)
		return nil, err
	}
	index_lookup := make(map[string]int)
	words := []Item{}
	slog.Info("Getting words from database...")
	for rows.Next() {
		var item Item

		err := rows.Scan(&item.Id, &item.Word, &item.Idf)
		if err != nil {
			slog.Error("Failed to scan", "error", err)
			return nil, err
		}
		index_lookup[item.Word] = item.Id - 1 //sql db uses 1 indexing
		words = append(words, item)
	}

	slog.Info("All words received, computing TF-IDF matrix")

	amount_of_words := len(words)
	var wg sync.WaitGroup
	ch := make(chan MatrixWrite)

	for i, article := range art_list {
		wg.Add(1)
		go func(i int, article *articles.Article) {
			defer wg.Done()
			words_map := make(map[string]int)
			total := 0
			for word := range strings.FieldsSeq(article.Stemmed) {
				total++
				_, ok := words_map[word]
				if !ok {
					words_map[word] = 1
				} else {
					words_map[word]++
				}
			}
			v_len := 0.0
			for word, count := range words_map {
				if _, ok := index_lookup[word]; ok {
					v_len += math.Pow(float64(count)/float64(total)*words[index_lookup[word]].Idf, 2)
				}
			}
			v_len = math.Sqrt(v_len)
			for word, count := range words_map {
				if _, ok := index_lookup[word]; ok {
					data := MatrixWrite{Row: index_lookup[word], Col: i, Val: float64(count) / float64(total) * words[index_lookup[word]].Idf / v_len}
					ch <- data
				}
			}
		}(i, article)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()
	sparseMatrix64 := make(map[int]map[int]float64)
	for data := range ch {
		if _, ok := sparseMatrix64[data.Row]; !ok {
			sparseMatrix64[data.Row] = make(map[int]float64)
		}
		sparseMatrix64[data.Row][data.Col] = data.Val
	}
	slog.Info("Computed TF-IDF matrix, calculating SVD...")
	svd := redsvd.NewGoRedSVD()
	svd.SetMatrix64(amount_of_words, len(art_list), sparseMatrix64) // Terms x Documents
	svd.RedSVD(K)
	U_old := svd.MatrixU()
	V := svd.MatrixV()
	S := svd.SingularValues()
	fmt.Println("V len", len(V), len(V[0]))
	fmt.Println("U len", len(U_old), len(U_old[0]))
	fmt.Println("S len", len(S))
	slog.Info("Computed SVD, beginning inserting vectors into database...")

	for j, article := range art_list {
		vec := []float32{}
		for k := range K {
			vec = append(vec, V[j][k]*S[k])
		}
		_, err := db.pool.Exec(context.Background(), "INSERT INTO documents (title, content, embedding) VALUES ($1, $2, $3)", article.Title, article.Content, pgvector.NewVector(vec))
		if err != nil {
			slog.Error("Failed to insert document data", "error", err)
			return nil, err
		}
	}

	slog.Info("Inserted vectors into database.")
	slog.Info("Saving U matrix to file...")

	U := ConvertToGonumMatrix(U_old)
	err = SaveMatrix("Umatrix.blob", U)
	if err != nil {
		slog.Error("Failed to save matrix", "error", err)
		return nil, err
	}

	slog.Info("Saved U matrix to file")

	return U, nil
}
