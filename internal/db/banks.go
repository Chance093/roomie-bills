package db

import (
	"fmt"
	"time"

	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

func (db *DB) AddHostedLink(roomie, linkToken string) error {
	// search roomie name in db and get roomie id
	sqlQuery := "SELECT id FROM roomies WHERE name = ?;"
	var roomieId int
	if err := db.QueryRow(sqlQuery, roomie).Scan(&roomieId); err != nil {
		return fmt.Errorf("Error querying roomie id and scanning row: %w", err)
	}

	// create bank record which saves roomie id and linkToken
	sqlInsert := "INSERT INTO banks(link_token, roomie_id, created_at, updated_at) VALUES(?, ?, ?, ?);"
	now := time.Now()
	if _, err := db.Exec(sqlInsert, linkToken, roomieId, now, now); err != nil {
		return fmt.Errorf("Error inserting roomie id and link token into banks table: %w", err)
	}

	return nil
}

func (db *DB) DeleteBankRecord(linkToken string) error {
	sqlStatement := "DELETE FROM banks WHERE link_token = ?;"
	_, err := db.Exec(sqlStatement, linkToken)
	if err != nil {
		return err
	}

	return nil
}

func (db *DB) UpdateBankRecord(linkToken, bankName string, accessToken plaid.AccessToken) (int, error) {
	// get bank id for func return and update statement
	sqlQuery := "SELECT id FROM banks WHERE link_token = ?;"
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

func (db *DB) GetBankAccessTokens() ([]string, error) {
	sqlQuery := "SELECT access_token FROM banks;"

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("Error querying account types: %w", err)
	}
	defer rows.Close()

	var accessTokens []string
	for rows.Next() {
		var accessToken string
		if err := rows.Scan(&accessToken); err != nil {
			return nil, fmt.Errorf("Failed to scan row for access tokens: %w", err)
		}

		accessTokens = append(accessTokens, accessToken)
	}

	if err = rows.Err(); err != nil {
		return accessTokens, fmt.Errorf("Error while iterating through rows: %w", err)
	}

	return accessTokens, nil
}
