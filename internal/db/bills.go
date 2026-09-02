package db

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

func (db *DB) AddBills(bills []plaid.Bill) ([]plaid.Bill, error) {
	newBills, err := db.findNewBills(bills)
	if err != nil {
		return nil, fmt.Errorf("Error while getting non existing bills: %w", err)
	}

	if len(newBills) < 1 {
		return newBills, nil
	}

	buildExecArgs := func() (string, []any) {
		var b strings.Builder
		b.WriteString("INSERT INTO bills (plaid_id, payee, date, total, account_id, created_at, updated_at) VALUES ")

		const argsPerRow = 7
		args := make([]any, argsPerRow*len(newBills))
		for i, bill := range newBills {
			b.WriteString("(?, ?, ?, ?, (SELECT id FROM accounts WHERE plaid_id = ?), ?, ?)")
			if i == len(newBills)-1 {
				b.WriteString(" ON CONFLICT(plaid_id) DO NOTHING;")
			} else {
				b.WriteString(", ")
			}

			t := time.Now()
			j := i * argsPerRow
			args[j] = bill.Id
			args[j+1] = bill.Payee
			args[j+2] = bill.Date
			args[j+3] = bill.Total
			args[j+4] = bill.AccountId
			args[j+5] = t
			args[j+6] = t
		}

		return b.String(), args
	}

	sqlInsert, sqlArgs := buildExecArgs()
	if _, err := db.Exec(sqlInsert, sqlArgs...); err != nil {
		return nil, fmt.Errorf("Error adding bills in db: %w", err)
	}

	return newBills, nil
}

func (db *DB) findNewBills(bills []plaid.Bill) ([]plaid.Bill, error) {
	var b strings.Builder
	b.WriteString("SELECT plaid_id FROM bills WHERE plaid_id IN (")

	// TODO: fix hints later
	for i, bill := range bills {
		if i == len(bills)-1 {
			b.WriteString(fmt.Sprintf("\"%s\");", bill.Id))
		} else {
			b.WriteString(fmt.Sprintf("\"%s\", ", bill.Id))
		}
	}

	rows, err := db.Query(b.String())
	if err != nil {
		return nil, fmt.Errorf("Error querying bill id's: %w", err)
	}
	defer rows.Close()

	// find bills that were already saved to db
	var oldBills []string
	for rows.Next() {
		var billId string
		if err := rows.Scan(&billId); err != nil {
			return nil, fmt.Errorf("Failed to scan row for bill id: %w", err)
		}

		oldBills = append(oldBills, billId)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Error while iterating through rows: %w", err)
	}

	// push new bills into slice
	var newBills []plaid.Bill
	for _, bill := range bills {
		if !slices.Contains(oldBills, bill.Id) {
			newBills = append(newBills, bill)
		}
	}

	return newBills, nil
}

func (db *DB) AddOwnerPayments(newBills []plaid.Bill) error {
	// build 1 sql query that inserts 4 payments per new bill
	var b strings.Builder
	b.WriteString("INSERT INTO payments (roomie_id, bill_id, status, created_at, updated_at) VALUES ")

	roomieIds, err := db.getRoomieIds()
	if err != nil {
		return fmt.Errorf("Error while querying roomie id's: %w", err)
	}

	const argsPerRow = 4
	paymentsPerBill := len(roomieIds)
	argsPerBill := argsPerRow * paymentsPerBill
	args := make([]any, argsPerBill*len(newBills))
	for i, bill := range newBills {
		for j, roomieId := range roomieIds {
			b.WriteString("(?, (SELECT id FROM bills where plaid_id = ?), \"Pending\",  ?, ?)")
			if i == len(newBills)-1 && j == len(roomieIds)-1 {
				b.WriteString(";")
			} else {
				b.WriteString(", ")
			}

			t := time.Now()
			k := (argsPerBill * i) + (argsPerRow * j)
			args[k] = roomieId
			args[k+1] = bill.Id
			args[k+2] = t
			args[k+3] = t
		}
	}

	if _, err := db.Exec(b.String(), args...); err != nil {
		return fmt.Errorf("Error adding payment in db: %w", err)
	}

	// update all payments by bill owners to be paid

	return nil
}

func (db *DB) getRoomieIds() ([]int64, error) {
	sqlQuery := "SELECT id FROM roomies;"

	rows, err := db.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("Error querying roomie id's: %w", err)
	}
	defer rows.Close()

	// find bills that were already saved to db
	var roomieIds []int64
	for rows.Next() {
		var roomieId int64
		if err := rows.Scan(&roomieId); err != nil {
			return nil, fmt.Errorf("Failed to scan row for roomie id: %w", err)
		}

		roomieIds = append(roomieIds, roomieId)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("Error while iterating through rows: %w", err)
	}

	return roomieIds, nil
}
