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

	pc.GetAccessToken("Chance")
}

