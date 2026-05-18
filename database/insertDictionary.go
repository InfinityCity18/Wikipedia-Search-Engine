package database

import (
	"context"
	"log/slog"
	"math"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/InfinityCity18/Wikipedia-Search-Engine/dict"
	"github.com/jackc/pgx/v5"
)

func (db *Database) insertDictionary(articles_list []*articles.Article) error {
	dictionary, length, err := dict.CreateDict(articles_list)
	if err != nil {
		slog.Error("Failed to create dictionary", "error", err)
	}
	entries := [][]any{}
	columns := []string{"word", "docs_count", "global_count", "idf"}
	tableName := "dictionary"

	for word, word_entry := range dictionary {
		entries = append(entries, []any{word, word_entry.Doc_count, word_entry.Global_count, math.Log(float64(length) / float64(word_entry.Doc_count))})
	}

	_, err = db.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(entries),
	)

	if err != nil {
		slog.Error("Failed to insert dictionary", "error", err)
		return nil
	}

	// _, err = db.pool.Exec(context.Background(),
	// 	fmt.Sprintf(
	// 		`DELETE FROM public.dictionary
	// 	WHERE id NOT IN (
	// 	SELECT id
	// 	FROM public.dictionary
	// 	ORDER BY idf DESC
	// 	LIMIT %v
	// 	);`, Limit_count))
	// if err != nil {
	// 	slog.Error("Failed to delete rows from dictionary", "error", err)
	// 	return nil
	// }
	// conn, err := db.pool.Acquire(context.Background())
	// if err != nil {
	// 	slog.Error("Connection acquire failed", "error", err)
	// 	return nil
	// }
	// _, err = conn.Exec(context.Background(),
	// 	`BEGIN;

	// 	WITH updated_rows AS (
	// 		SELECT
	// 			id,
	// 			(row_number() OVER (ORDER BY id)) * -1 AS temp_id
	// 		FROM dictionary
	// 	)
	// 	UPDATE dictionary t
	// 	SET id = u.temp_id
	// 	FROM updated_rows u
	// 	WHERE t.id = u.id;

	// 	UPDATE dictionary
	// 	SET id = id * -1
	// 	WHERE id < 0;

	// 	COMMIT;`)
	// if err != nil {
	// 	slog.Error("Failed to reindex", "error", err)
	// 	return nil
	// }
	return nil
}
