package database

import (
	"context"
	"log/slog"
	"math"
	"strings"
	"sync"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/InfinityCity18/Wikipedia-Search-Engine/pythonsvd"
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

const (
	k1 = 1.5
	b  = 0.75
)

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

	var totalWordsAllDocs int
	for _, article := range art_list {
		totalWordsAllDocs += len(strings.Fields(article.Stemmed))
	}
	avgdl := float64(totalWordsAllDocs) / float64(len(art_list))

	for i, article := range art_list {
		wg.Add(1)
		go func(i int, article *articles.Article) {
			defer wg.Done()
			words_map := make(map[string]int)
			docLen := len(strings.Fields(article.Stemmed))
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
					data := MatrixWrite{Row: index_lookup[word], Col: i, Val: computeBM25Weight(count, float64(docLen), avgdl, words[index_lookup[word]].Idf)}
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
	U, S, V, err := pythonsvd.TruncatedSVD(amount_of_words, len(art_list), K, sparseMatrix64)
	slog.Info("Computed SVD, beginning inserting vectors into database...")
	for j, article := range art_list {
		vec := []float32{}
		for k := range K {
			vec = append(vec, float32(V.At(k, j)*S.At(k, k)))
		}
		_, err := db.pool.Exec(context.Background(), "INSERT INTO documents (title, content, embedding) VALUES ($1, $2, $3)", article.Title, article.Content, pgvector.NewVector(vec))
		if err != nil {
			slog.Error("Failed to insert document data", "error", err)
			return nil, err
		}
	}

	slog.Info("Inserted vectors into database.")
	slog.Info("Saving U matrix to file...")

	err = SaveMatrix("Umatrix.blob", U)
	if err != nil {
		slog.Error("Failed to save matrix", "error", err)
		return nil, err
	}

	slog.Info("Saved U matrix to file")

	return U, nil
}

func computeTFIDFWeight(count int, total int, idf float64, vLen float64) float64 {
	if vLen == 0 {
		return 0
	}
	return (float64(count) / float64(total)) * idf / vLen
}

func computeBM25Weight(count int, docLen float64, avgdl float64, idf float64) float64 {
	tf := float64(count)
	numerator := tf * (k1 + 1.0)
	denominator := tf + k1*(1.0-b+b*(docLen/avgdl))
	return idf * (numerator / denominator)
}
