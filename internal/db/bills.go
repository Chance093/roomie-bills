package db

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

type Bill struct {
	Id     int64
	Payee  string
	Date   string
	Total  float64
	Roomie string
}

func (db *DB) FindNewPlaidBills(bills []plaid.Bill) ([]plaid.Bill, error) {
	// query all bill plaid_id's to see which already exist
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

	// create slice containing bills that have already been saved
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

	// create new bills by comparing bills with the old bills slice
	var newBills []plaid.Bill
	for _, bill := range bills {
		if !slices.Contains(oldBills, bill.Id) {
			newBills = append(newBills, bill)
		}
	}

	return newBills, nil
}

func (db *DB) AddBillsAndPayments(bills []plaid.Bill) ([]Bill, error) {
	// begin a transaction
	tx, err := db.BeginTx(context.TODO(), nil)
	if err != nil {
		return nil, fmt.Errorf("Failed to start db transaction: %w", err)
	}
	defer tx.Rollback()

	// get all roomie id's
	roomieIds, err := getRoomieIds(tx)
	if err != nil {
		return nil, fmt.Errorf("Failed to get roomie id's for bill insertion: %w", err)
	}

	// save bills to db and add initial payments associated with each bill
	savedBills := make([]Bill, len(bills))
	for i, bill := range bills {
		id, err := addBill(tx, bill)
		if err != nil {
			return nil, fmt.Errorf("Failed to add bill to db: %w", err)
		}

		roomie, err := getRoomieWhoPaidBill(tx, id)
		if err != nil {
			return nil, fmt.Errorf("Failed to get roomie while adding bill: %w", err)
		}

		if err := addPaymentsForBill(tx, bill.Id, roomie, roomieIds); err != nil {
			return nil, fmt.Errorf("Failed to add payments associated with bill %d: %w", id, err)
		}

		savedBills[i] = Bill{
			Id:     id,
			Payee:  bill.Payee,
			Date:   bill.Date,
			Total:  bill.Total,
			Roomie: roomie.name,
		}
	}

	// commit all insertions
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("Failed to commit transactions: %w", err)
	}

	return savedBills, nil
}

func getRoomieIds(tx *sql.Tx) ([]int64, error) {
	sqlQuery := "SELECT id FROM roomies;"
	rows, err := tx.Query(sqlQuery)
	if err != nil {
		return nil, fmt.Errorf("Error querying roomie id's: %w", err)
	}
	defer rows.Close()

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

func addBill(tx *sql.Tx, bill plaid.Bill) (int64, error) {
	sqlInsert := `
	INSERT INTO bills(plaid_id, payee, date, total, account_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, (SELECT id FROM accounts WHERE plaid_id = ?), ?, ?);
	`

	t := time.Now()
	res1, err := tx.Exec(sqlInsert, bill.Id, bill.Payee, bill.Date, bill.Total, bill.AccountId, t, t)
	if err != nil {
		return 0, fmt.Errorf("Failed to insert bill in db: %w", err)
	}

	id, err := res1.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("Failed to get last insert id: %w", err)
	}

	return id, nil
}

type Roomie struct {
	id   int64
	name string
}

func getRoomieWhoPaidBill(tx *sql.Tx, billId int64) (Roomie, error) {
	sqlQuery := `
	SELECT roomies.id, roomies.name FROM bills
	INNER JOIN accounts ON bills.account_id = accounts.id
	INNER JOIN banks ON accounts.bank_id = banks.id
	INNER JOIN roomies ON banks.roomie_id = roomies.id
	WHERE bills.id = ?;
	`

	var roomie Roomie
	if err := tx.QueryRow(sqlQuery, billId).Scan(&roomie.id, &roomie.name); err != nil {
		return Roomie{}, fmt.Errorf("Failed to query roomie associated with bill %d: %w", billId, err)
	}

	return roomie, nil
}

func addPaymentsForBill(tx *sql.Tx, billPlaidId string, roomie Roomie, roomieIds []int64) error {
	var b strings.Builder
	b.WriteString("INSERT INTO payments (roomie_id, bill_id, status, created_at, updated_at) VALUES ")

	const argsPerRow = 5
	args := make([]any, argsPerRow*len(roomieIds))
	for i, roomieId := range roomieIds {
		b.WriteString("(?, (SELECT id FROM bills where plaid_id = ?), ?, ?, ?)")
		if i == len(roomieIds)-1 {
			b.WriteString(";")
		} else {
			b.WriteString(", ")
		}

		t := time.Now()
		j := i * argsPerRow

		args[j] = roomieId
		args[j+1] = billPlaidId
		if roomieId == roomie.id {
			args[j+2] = "Paid"
		} else {
			args[j+2] = "Pending"
		}
		args[j+3] = t
		args[j+4] = t
	}

	if _, err := tx.Exec(b.String(), args...); err != nil {
		return fmt.Errorf("Error adding payment in db: %w", err)
	}

	return nil
}
