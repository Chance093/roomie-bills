package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/utils"
)

func main() {
	// get env vars
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	// get roomie name
	roomie, err := getRoomieName(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	// get hosted link from plaid
	pc := lib.NewPlaidClient(env)
	hostedLink, err := pc.GetHostedLink(roomie, env)
	if err != nil {
		log.Fatal(err)
	}

	// save hosted link info to db
	db := db.NewDB()
	defer db.Close()
	if err := db.AddHostedLink(roomie, hostedLink.LinkToken); err != nil {
		log.Fatal(err)
	}

	// send url to discord channel
	dc, err := lib.NewDiscordClient(env)
	if err != nil {
		log.Fatal(err)
	}

	if err := dc.SendHostedLink(roomie, hostedLink.Url); err != nil {
		fmt.Println("Failed to send host link. Deleting bank record from db...")

		if err := db.DeleteBankRecord(hostedLink.LinkToken); err != nil {
			log.Fatal(err)
		}

		log.Fatal(err)
	}
}

func getRoomieName(stdin *os.File) (string, error) {
	scanner := bufio.NewScanner(stdin)
	fmt.Println("Which roomie is this hosted link for?")
	if scanner.Scan() {
		raw := scanner.Text()
		roomie, err := utils.ParseRoomieName(raw)
		if err != nil {
			return "", err
		}
		fmt.Println("") // new line to split up stdout

		return roomie, nil
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("Error reading input: %w", err)
	}

	return "", errors.New("This part of function unreachable")
}

