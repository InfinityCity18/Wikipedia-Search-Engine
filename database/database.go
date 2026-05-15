package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	pool *pgxpool.Pool
}

func New(connString string) (Database, error) {
	pool, err := pgxpool.New(context.Background(), connString)
	pool.Exec()
	pool.Query()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return Database{}, err
	}
	_, err = pool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS dictionary (
		id integer AUTO_INCREMENT PRIMARY KEY,
		word VARCHAR(50) NOT NULL,
		docs_count INT,
		global_count INT
		);`)
	if err != nil {
		slog.Error("Failed to execute init dictionary table query", "error", err)
		return Database{}, err
	}
	_, err = pool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS documents (
		id integer AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(100) NOT NULL,
		content text NOT NULL,
		
		);`)
	if err != nil {
		slog.Error("Failed to execute init dictionary table query", "error", err)
		return Database{}, err
	}
	_, err = pool.Exec(context.Background(), "CREATE EXTENSION IF NOT EXISTS vector")
	if err != nil {
		slog.Error("Failed to create vector extension", "error", err)
		return Database{}, err
	}


	return Database{pool}, nil
}

func (d *Database) Close() {
	d.Close()
}
