package db

import (
	"database/sql"
	"fmt"
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

func (db *DB) AddHostedLink(roomie, linkToken string) error {
	// search roomie name in db and get roomie id
	sqlQuery := `SELECT id FROM roomies WHERE name = ?;`
	var roomieId int
	if err := db.QueryRow(sqlQuery, roomie).Scan(&roomieId); err != nil {
		return fmt.Errorf("Error querying roomie id and scanning row: %w", err)
	}

	// create bank record which saves roomie id and linkToken
	sqlInsert := `INSERT INTO banks(link_token, roomie_id) VALUES(?, ?);`
	if _, err := db.Exec(sqlInsert, linkToken, roomieId); err != nil {
		return fmt.Errorf("Error inserting roomie id and link token into banks table: %w", err)
	}

	fmt.Println("Link token and hosted link saved to db")

	return nil
}

func (db *DB) DeleteBankRecord(linkToken string) error {
	sqlStatement := `DELETE FROM banks WHERE link_token = ?;`
	_, err := db.Exec(sqlStatement, linkToken)
	if err != nil {
		return err
	}

	fmt.Println("deleted bank record")

	return nil
}

func (db *DB) UpdateBankRecord(linkToken, accessToken, plaidId, bankName string) error {
	sqlStatement := `
	UPDATE banks 
	SET access_token = ? plaid_id = ? access_token = ?
	WHERE link_token = ?
	`

	_, err := db.Exec(sqlStatement, accessToken, plaidId, bankName, linkToken)
	if err != nil {
		return err
	}

	fmt.Println("updated bank record")

	return nil
}
