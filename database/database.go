package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
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
	row := db.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM dictionary")
	var s int
	row.Scan(&s)
	if s == 0 {
		art_list, err := db.insertDictionary(dataDirPath)
		if err != nil {
			slog.Error("Failed to insert dictionary", "error", err)
			return Database{}, err
		}
		db.computeTfidf(art_list)
	}
	return db, err
}

func (d *Database) Close() {
	d.Close()
}
