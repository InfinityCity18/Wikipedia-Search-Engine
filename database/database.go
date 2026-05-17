package database

import (
	"context"
	"log/slog"

	"github.com/InfinityCity18/Wikipedia-Search-Engine/articles"
	"github.com/InfinityCity18/Wikipedia-Search-Engine/dict"
	"github.com/jackc/pgx/v5/pgxpool"
	"gonum.org/v1/gonum/mat"
)

type Database struct {
	pool    *pgxpool.Pool
	umatrix *mat.Dense
}

const K = 512

func Create(connString string, dataDirPath string) (Database, error) {
	db, err := initalize(connString)
	if err != nil {
		slog.Error("Failed to initalize database", "error", err)
		return Database{}, err
	}
	slog.Info("Database initalized")

	row := db.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM dictionary")
	var s int
	row.Scan(&s)
	var art_list []*articles.Article
	if s == 0 {
		art_list, err = dict.CreateArticlesList(dataDirPath)
		slog.Info("Created list of articles")
		if err != nil {
			slog.Error("Failed to create articles list", "error", err)
			return Database{}, err
		}
		slog.Info("Created list of articles")
		err = db.insertDictionary(art_list)
		if err != nil {
			slog.Error("Failed to insert dictionary", "error", err)
			return Database{}, err
		}
		slog.Info("Inserted dictionary into database")
	}

	row = db.pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM documents")
	row.Scan(&s)
	if s == 0 {
		if art_list == nil {
			art_list, err = dict.CreateArticlesList(dataDirPath)
			if err != nil {
				slog.Error("Failed to create articles list", "error", err)
				return Database{}, err
			}
			slog.Info("Created list of articles")
		}
		db.umatrix, err = db.insertDocuments(art_list)
		if err != nil {
			slog.Error("Failed to insert documents", "error", err)
			return Database{}, nil
		}
	}
	if db.umatrix == nil {
		db.umatrix, err = LoadMatrix("Umatrix.blob")
		if err != nil {
			slog.Error("Failed to load U matrix from file", "error", err)
			return Database{}, err
		}
	}
	slog.Info("Database fully prepared.")
	return db, err
}

func (d *Database) Close() {
	d.Close()
}
