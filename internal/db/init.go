package db

import (
	"database/sql"
	"time"
)

type AccountSubtype string

const (
	ACCOUNTSUBTYPE_CHECKING    AccountSubtype = "checking"
	ACCOUNTSUBTYPE_CREDIT_CARD AccountSubtype = "credit card"
	ACCOUNTSUBTYPE_SAVINGS     AccountSubtype = "savings"
)

var accountTypes = []AccountSubtype{
	ACCOUNTSUBTYPE_CHECKING, ACCOUNTSUBTYPE_CREDIT_CARD, ACCOUNTSUBTYPE_SAVINGS,
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
		created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS banks (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		plaid_id TEXT UNIQUE,
		link_token TEXT NOT NULL UNIQUE,
		name TEXT,
		access_token TEXT UNIQUE,
		roomie_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
	);

  CREATE TABLE IF NOT EXISTS account_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
  );

	CREATE TABLE IF NOT EXISTS accounts (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		plaid_id TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		type_id INTEGER NOT NULL,
		bank_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS bills (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		payee TEXT NOT NULL,
		date DATETIME NOT NULL,
		total INTEGER NOT NULL,
		account_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
  );

	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		roomie_id INTEGER NOT NULL,
		bill_id INTEGER NOT NULL,
		created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
	);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	buildExecArgs := func(types []AccountSubtype) (string, []any) {
		sqlInsert := `
		INSERT OR IGNORE INTO account_types (name, created_at, updated_at)
		VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?);
		`
		const argsPerRow = 3
		args := make([]any, argsPerRow*len(types))

		for i, v := range types {
			j := i * argsPerRow
			args[j] = v
			args[j+1] = time.Now()
			args[j+2] = time.Now()
		}

		return sqlInsert, args
	}

	sqlInsert, sqlArgs := buildExecArgs(accountTypes)
	_, err = db.Exec(sqlInsert, sqlArgs...)
	if err != nil {
		return nil, err
	}

	createRoomies := `
	INSERT OR IGNORE INTO roomies (name, created_at, updated_at)
	VALUES ("Chance", ?, ?), ("Kane", ?, ?), ("Alex", ?, ?), ("Madison", ?, ?);
	`
	roomiesArgs := make([]any, 8)
	for i := range 8 {
		roomiesArgs[i] = time.Now()
	}
	_, err = db.Exec(createRoomies, roomiesArgs...)
	if err != nil {
		return nil, err
	}

	return db, nil
}
