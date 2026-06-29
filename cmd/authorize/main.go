package main

import (
	"log"
	"os"

	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	plaidClientId := os.Getenv("PLAID_CLIENT_ID")
	plaidSecretKey := os.Getenv("PLAID_SANDBOX_SECRET")

	pc := lib.NewPlaidClient(plaidClientId, plaidSecretKey)

	// grab name of roomie from input

	// create plaid link and send to roomie (either sms or url that is pasted to discord)
	hostedLink, err := pc.GetHostedLink("Kane")
	if err != nil {
		log.Fatal(err)
	}

	// save hosted link info to db

	// send url to discord channel
}
