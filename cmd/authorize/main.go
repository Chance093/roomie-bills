package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/joho/godotenv"
)

func main() {
	// get env vars
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	plaidClientId := os.Getenv("PLAID_CLIENT_ID")
	plaidSecretKey := os.Getenv("PLAID_SANDBOX_SECRET")

	// get name of roomie
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Who is this link for?")
	if scanner.Scan() {
		roomie := scanner.Text()

		// get hosted link from plaid
		pc := lib.NewPlaidClient(plaidClientId, plaidSecretKey)
		hostedLink, err := pc.GetHostedLink(roomie)
		if err != nil {
			log.Fatal(err)
		}

		// save hosted link info to db
		db := db.NewDB()
		db.AddHostedLink(roomie, hostedLink.LinkToken) 

		// send url to discord channel
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}
}
