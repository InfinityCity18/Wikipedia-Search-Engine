package database

import (
	"context"
	"log/slog"
	"math"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/InfinityCity18/Wikipedia-Search-Engine/dict"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

type Database struct {
	pool *pgxpool.Pool
}

func Create(connString string, dataDirPath string) (Database, error) {
	db, err := initalize(connString)
	if err != nil {
		slog.Error("Failed to initalize database", "error", err)
		return Database{}, nil
	}
	art_list, err := db.insertDictionary(dataDirPath)
	if err != nil {
		slog.Error("Failed to insert dictionary", "error", err)
		return Database{}, err
	}
	return db, err
}

func initalize(connString string) (Database, error) {
	conf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		slog.Error("Failed to parse config", "error", err)
		return Database{}, err
	}
	conf.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}
	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return Database{}, err
	}
	_, err = pool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS dictionary (
		id SERIAL PRIMARY KEY,
		word VARCHAR(100) NOT NULL,
		docs_count INT,
		global_count INT,
		idf double precision
		);`)
	if err != nil {
		slog.Error("Failed to execute init dictionary table query", "error", err)
		return Database{}, err
	}
	_, err = pool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS documents (
		id SERIAL PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		content TEXT NOT NULL
		);`)
	if err != nil {
		slog.Error("Failed to execute init documents table query", "error", err)
		return Database{}, err
	}
	_, err = pool.Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		slog.Error("Failed to create vector extension", "error", err)
		return Database{}, err
	}

	return Database{pool}, nil
}

func (db *Database) insertDictionary(dataDirPath string) ([]*articles.Article, error) {
	dictionary, articles_list, length, err := dict.CreateDict(dataDirPath)
	if err != nil {
		slog.Error("Failed to create dictionary", "error", err)
	}
	entries := [][]any{}
	columns := []string{"word", "docs_count", "global_count", "idf"}
	tableName := "dictionary"

	for word, word_entry := range dictionary {
		entries = append(entries, []any{word, word_entry.Doc_count, word_entry.Global_count, float64(word_entry.Global_count) * math.Log(float64(length)/float64(word_entry.Doc_count))})
	}

	_, err = db.pool.CopyFrom(
		context.Background(),
		pgx.Identifier{tableName},
		columns,
		pgx.CopyFromRows(entries),
	)

	if err != nil {
		slog.Error("Failed to insert dictionary", "error", err)
		return nil, err
	}

	_, err = db.pool.Exec(context.Background(),
		`DELETE FROM public.dictionary
		WHERE id NOT IN (
		SELECT id
		FROM public.dictionary
		ORDER BY idf DESC
		LIMIT 500000
		);`)
	if err != nil {
		slog.Error("Failed to delete rows from dictionary", "error", err)
		return nil, err
	}
	return articles_list, nil
}

func (d *Database) Close() {
	d.Close()
}
