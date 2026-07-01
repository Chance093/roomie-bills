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

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./test.db")
	if err != nil {
		return nil, err
	}
	defer db.Close()

	sqlStmt := `
	CREATE TABLE IF NOT EXISTS roomies (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS banks (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		plaid_id TEXT UNIQUE,
		link_token TEXT NOT NULL UNIQUE,
		name TEXT,
		access_token TEXT UNIQUE,
		roomie_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

  CREATE TABLE IF NOT EXISTS account_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );

	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		plaid_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		type_id INTEGER NOT NULL,
		bank_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS bills (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		payee TEXT NOT NULL,
		date DATETIME NOT NULL,
		total INTEGER NOT NULL,
		account_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
  );`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	createAccountTypes := `
  INSERT OR IGNORE INTO account_types (name)
  VALUES ("Checking"), ("Savings"), ("Credit");
  `
	_, err = db.Exec(createAccountTypes)
	if err != nil {
		return nil, err
	}

	createRoomies := `
	INSERT OR IGNORE INTO roomies (name)
	VALUES ("Chance"), ("Kane"), ("Alex"), ("Madison");
	`
	_, err = db.Exec(createRoomies)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (db *DB) AddHostedLink(roomie, linkToken string) {
	// search roomie name in db and get roomie id

	// create bank record which saves roomie id and linkToken
}
