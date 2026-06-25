package crons

func CheckForNewBillsCron() {
	// request transactions from plaid api
	// find all transactions that are NEW bills
	// split bills 4 ways
	// send discord message
	// save new transactions to db
}

func EndOfMonthSummaryCron() {
	// check db for unpaid bills
	// send discord message
}
