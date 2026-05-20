package database

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

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
	_, err = pool.Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		slog.Error("Failed to create vector extension", "error", err)
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
		fmt.Sprintf(
			`CREATE TABLE IF NOT EXISTS documents (
		id SERIAL PRIMARY KEY,
		title VARCHAR(1000) NOT NULL,
		content TEXT NOT NULL,
		embedding vector(%v)
		);`, K))
	if err != nil {
		slog.Error("Failed to execute init documents table query", "error", err)
		return Database{}, err
	}

	_, err = pool.Exec(context.Background(), "CREATE INDEX IF NOT EXISTS hnsw_vec_index ON documents USING hnsw (embedding vector_cosine_ops);")
	if err != nil {
		slog.Error("Failed to create index on embedding vector", "error", err)
		return Database{}, err
	}

	return Database{pool, nil}, nil
}
