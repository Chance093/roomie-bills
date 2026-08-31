package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

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
	sqlQuery := "SELECT id, name FROM account_types;"

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
