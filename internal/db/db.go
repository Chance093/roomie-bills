package db

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

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

	createAccountTypes := `
  INSERT OR IGNORE INTO account_types (name, created_at, updated_at)
  VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?);
  `

	// converts slice of AccountSubtype to slice of any for variadic db.Exec()
	convertToArgs := func(types []AccountSubtype) []any {
		res := make([]any, len(types)*3)

		for i, v := range types {
			res[i*3] = v
			res[i*3+1] = time.Now()
			res[i*3+2] = time.Now()
		}

		return res
	}

	_, err = db.Exec(createAccountTypes, convertToArgs(accountTypes)...)
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

func (db *DB) AddHostedLink(roomie, linkToken string) error {
	// search roomie name in db and get roomie id
	sqlQuery := `SELECT id FROM roomies WHERE name = ?;`
	var roomieId int
	if err := db.QueryRow(sqlQuery, roomie).Scan(&roomieId); err != nil {
		return fmt.Errorf("Error querying roomie id and scanning row: %w", err)
	}

	// create bank record which saves roomie id and linkToken
	sqlInsert := `INSERT INTO banks(link_token, roomie_id, created_at, updated_at) VALUES(?, ?, ?, ?);`
	now := time.Now()
	if _, err := db.Exec(sqlInsert, linkToken, roomieId, now, now); err != nil {
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
	// get bank id for func return and update statement
	sqlQuery := `
	SELECT id FROM banks WHERE link_token = ?;
	`
	var bankId int
	// NOTE: this query intentionally comes first before mutation
	if err := db.QueryRow(sqlQuery, linkToken).Scan(&bankId); err != nil {
		return 0, fmt.Errorf("Error querying bank id and scanning row: %w", err)
	}

	sqlStatement := `
	UPDATE banks 
	SET access_token = ?, plaid_id = ?, name = ?, updated_at = ?
	WHERE id = ?;
	`

	if _, err := db.Exec(sqlStatement, accessToken.Token, accessToken.ItemId, bankName, time.Now(), bankId); err != nil {
		return 0, fmt.Errorf("Error while updating bank record: %w", err)
	}

	return bankId, nil
}

func (db *DB) AddAccounts(accounts []plaid.Account, bankId int) error {
	// get account types and type id's
	typesMap, err := db.getAccountTypes()
	if err != nil {
		return fmt.Errorf("Could not get account types map: %w", err)
	}

	buildExecArgs := func() (string, []any) {
		var b strings.Builder
		b.WriteString("INSERT INTO accounts (plaid_id, name, type_id, bank_id, created_at, updated_at) VALUES ")

		const argsPerRow = 6
		args := make([]any, argsPerRow*len(accounts))
		for i, acc := range accounts {
			b.WriteString("(?, ?, ?, ?, ?, ?)")
			if i == len(accounts)-1 {
				b.WriteString(";")
			} else {
				b.WriteString(", ")
			}

			t := time.Now()
			j := i * argsPerRow
			args[j] = acc.PlaidId
			args[j+1] = acc.Name
			args[j+2] = typesMap[acc.Type]
			args[j+3] = bankId
			args[j+4] = t
			args[j+5] = t
		}

		return b.String(), args
	}

	sqlInsert, sqlArgs := buildExecArgs()
	if _, err := db.Exec(sqlInsert, sqlArgs...); err != nil {
		return fmt.Errorf("Error adding account in db: %w", err)
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
