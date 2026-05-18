package database

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"strings"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/pgvector/pgvector-go"
	"gonum.org/v1/gonum/mat"
)

func (db *Database) ProcessQuery(query string) ([]*articles.Article, error) {
	query, err := articles.StemAndRemoveStopWordsString(query)
	if err != nil {
		slog.Error("Failed to stem and remove stop words of query", "error", err)
		return nil, err
	}
	ids := db.getStemmedWordsId(strings.Fields(query))
	pgv := pgvector.NewVector(multipliedQuery(ids, db.umatrix))
	fmt.Println(pgv)
	rows, err := db.pool.Query(
		context.Background(),
		"SELECT title,content FROM documents WHERE LENGTH(content) != 0 ORDER BY (1 - (embedding <=> $1)) DESC LIMIT 5", pgv)
	if err != nil {
		slog.Error("Failed to get relevant articles", "error", err)
		return nil, err
	}
	art_list := []*articles.Article{}
	for rows.Next() {
		var title, content string
		err := rows.Scan(&title, &content)
		if err != nil {
			slog.Error("Failed to scan from relevant article", "error", err)
			continue
		}
		art := articles.Article{Title: title, Content: content}
		art_list = append(art_list, &art)
	}
	return art_list, nil
}

func (db *Database) getStemmedWordsId(words []string) map[int]float32 {
	ids_and_idfs := make(map[int]float32)
	for _, word := range words {
		row := db.pool.QueryRow(context.Background(),
			"SELECT id, idf FROM dictionary WHERE word = $1", word)
		var id int
		var idf float32
		if err := row.Scan(&id, &idf); err != nil {
			slog.Error("Failed to scan row", "error", err)
		} else {
			ids_and_idfs[id-1] = idf
		}
	}
	return ids_and_idfs
}

func multipliedQuery(query map[int]float32, U *mat.Dense) []float32 {
	final_vector := make([]float32, K)
	normalize(query)
	for k := range K {
		var sum float32 = 0.0
		for id, idf := range query {
			sum += float32(U.At(id, k)) * idf
		}
		final_vector[k] = float32(sum)
	}
	magnitude := 0.0
	for _, val := range final_vector {
		magnitude += math.Pow(float64(val), 2.0)
	}
	magnitude = math.Sqrt(magnitude)
	for i := range final_vector {
		final_vector[i] /= float32(magnitude)
	}
	return final_vector
}

func normalize(v map[int]float32) {
	var sum float32 = 0.0
	for _, val := range v {
		sum += float32(math.Pow(float64(val), 2))
	}
	sum = float32(math.Sqrt(float64(sum)))
	for k := range v {
		v[k] /= sum
	}
}
