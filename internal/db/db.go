package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewDB() *DB {
	db, err := initDB()
	if err != nil {
		log.Fatal(err)
	}

	return &DB{db}
}
