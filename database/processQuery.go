package database

import (
	"log/slog"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"gonum.org/v1/gonum/mat"
)

func (db *Database) ProcessQuery(query string) (*articles.Article, error) {
	query, err := articles.StemAndRemoveStopWordsString(query)
	if err != nil {
		slog.Error("Failed to stem and remove stop words of query", "error", err)
		return nil, err
	}

}

func (db *Database) getStemmedWordsId(words []string) []int {
	ids := []int{}
	for word := range words {

	}
	return ids
}

// query is list of ids of words that are in query string
func multipliedQuery(query []int, U *mat.Dense) []float32 {
	final_vector := make([]float32, K, K)
	for k := range K {
		sum := 0.0
		for id := range query {
			sum += U.At(id, k)
		}
		final_vector[k] = float32(sum)
	}
	return final_vector
}
