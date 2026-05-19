package database

import (
	"context"
	"log/slog"
	"math"
	"strings"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/pgvector/pgvector-go"
	"gonum.org/v1/gonum/mat"
)

const arts_limit = 10

func (db *Database) ProcessQuery(query string) ([]*articles.Article, error) {
	query, err := articles.StemAndRemoveStopWordsString(query)
	if err != nil {
		slog.Error("Failed to stem and remove stop words of query", "error", err)
		return nil, err
	}
	ids := db.getStemmedWordsId(strings.Fields(query))
	pgv := pgvector.NewVector(multipliedQuery(ids, db.umatrix))
	rows, err := db.pool.Query(
		context.Background(),
		"SELECT title,content,(1 - (embedding <=> $1)) FROM documents WHERE LENGTH(content) != 0 ORDER BY (1 - (embedding <=> $1)) DESC LIMIT $2", pgv, arts_limit)
	if err != nil {
		slog.Error("Failed to get relevant articles", "error", err)
		return nil, err
	}
	art_list := []*articles.Article{}
	for rows.Next() {
		var title, content string
		var sim float32
		err := rows.Scan(&title, &content, &sim)
		if err != nil {
			slog.Error("Failed to scan from relevant article", "error", err)
			continue
		}
		art := articles.Article{Title: title, Content: content, Sim: sim}
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
			sum += float32(U.At(id, k)) * idf * ((1.0 * (1.5 + 1.0)) / (1.0 + 1.5))
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
