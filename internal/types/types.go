package types

import "github.com/Chance093/roomie-bills/internal/db"

type SplitBill struct {
	db.Bill
	Split float64
}
