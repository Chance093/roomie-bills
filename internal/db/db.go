package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/Chance093/roomie-bills/internal/lib/plaid"
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
  );

	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		roomie_id INTEGER NOT NULL,
		bill_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	createAccountTypes := `
  INSERT OR IGNORE INTO account_types (name)
  VALUES (?), (?), (?);
  `

	// converts slice of AccountSubtype to slice of any for variadic db.Exec()
	convertToAny := func(types []AccountSubtype) []any {
		res := make([]any, len(types))

		for i, v := range types {
			res[i] = v
		}

		return res
	}

	_, err = db.Exec(createAccountTypes, convertToAny(accountTypes)...)
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

func (db *DB) AccessTokenExists(accessToken string) (bool, error) {
	sqlStatement := `
	SELECT count(access_token) FROM banks WHERE access_token = ?;
	`

	query, err := db.Prepare(sqlStatement)
	if err != nil {
		return false, fmt.Errorf("Error while making db query: %w", err)
	}
	defer query.Close()

	var count string
	err = query.QueryRow(accessToken).Scan(&count)

	switch {
	case err == sql.ErrNoRows || count == "0":
		return false, nil
	case err != nil:
		return false, fmt.Errorf("Error while counting rows: %w", err)
	default:
		return true, nil
	}
}

func (db *DB) UpdateBankRecord(linkToken, bankName string, accessToken plaid.AccessToken) (int, error) {
	// TODO: use current timestamp for updated_at
	sqlStatement := `
	UPDATE banks 
	SET access_token = ?, plaid_id = ?, name = ?
	WHERE link_token = ?;
	`

	if _, err := db.Exec(sqlStatement, accessToken.Token, accessToken.ItemId, bankName, linkToken); err != nil {
		return 0, err
	}

	// get bank id of recently updated bank
	sqlQuery := `
	SELECT id FROM banks WHERE access_token = ?;
	`
	var bankId int
	if err := db.QueryRow(sqlQuery, accessToken.Token).Scan(&bankId); err != nil {
		return 0, fmt.Errorf("Error querying bank id and scanning row: %w", err)
	}

	return bankId, nil
}

func (db *DB) AddAccounts(accounts []plaid.Account, bankId int) error {
	// get account types and type id's
	typesMap, err := db.getAccountTypes()
	if err != nil {
		return fmt.Errorf("Could not get account types map: %w", err)
	}

	sqlInsert := `
	INSERT INTO accounts (plaid_id, name, type_id, bank_id) VALUES (?, ?, ?, ?);
	`

	// TODO: make this concurrent
	for _, acc := range accounts {
		if _, err := db.Exec(sqlInsert, acc.PlaidId, acc.Name, typesMap[acc.Type], bankId); err != nil {
			return fmt.Errorf("Error adding accounts in db: %w", err)
		}
	}

	return nil
}

type AccountType struct {
	Id   int
	Name string
}

type AccountTypeToId map[string]int

func (db *DB) getAccountTypes() (AccountTypeToId, error) {
	sqlQuery := `
	SELECT id, name FROM account_types;
	`

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("Error querying account types: %w", err)
	}
	defer rows.Close()

	m := make(AccountTypeToId)
	for rows.Next() {
		var at AccountType
		if err := rows.Scan(&at.Id, &at.Name); err != nil {
			return nil, fmt.Errorf("Failed to scan row: %w", err) // TODO: better error handling
		}

		m[at.Name] = at.Id
	}

	if err = rows.Err(); err != nil {
		return m, fmt.Errorf("Error while iterating through rows: %w", err)
	}

	return m, nil
}
