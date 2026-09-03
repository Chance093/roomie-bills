package types

import "github.com/Chance093/roomie-bills/internal/lib/plaid"

type SplitBill struct {
	plaid.Bill
	Split float64
}
