package main

import (
	"context"
	"log"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/discord"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/lib/taskqueue"
	"github.com/Chance093/roomie-bills/internal/tasks"
)

func main() {
	// init everything
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatalf("Could not get env variables: %s\n", err.Error())
	}

	plaidClient := plaid.NewClient(env)
	discordClient, err := discord.NewClient(env)
	if err != nil {
		log.Fatalf("Could not connect to discord client: %s\n", err.Error())
	}
	db := db.NewDB()
	defer db.Close()

	ctx := context.Background()
	taskClient := taskqueue.NewClient(ctx, taskqueue.ClientOpts{})

	// config server and handlers
	srv := taskqueue.NewServer(ctx, taskqueue.ServerOpts{})
	mux := taskqueue.NewServeMux()
	handler := tasks.NewHandler(plaidClient, taskClient, discordClient, db)

	mux.HandleFunc(tasks.TypeGetAccessToken, handler.GetAccessToken)
	mux.HandleFunc(tasks.TypeGetBank, handler.GetBankName)
	mux.HandleFunc(tasks.TypeUpdateBank, handler.UpdateBank)
	mux.HandleFunc(tasks.TypeGetAccounts, handler.GetAccounts)
	mux.HandleFunc(tasks.TypeAddAccounts, handler.AddAccounts)
	mux.HandleFunc(tasks.TypeGetAccessTokens, handler.GetAccessTokens)
	mux.HandleFunc(tasks.TypeGetBills, handler.GetBills)
	mux.HandleFunc(tasks.TypeGetNewBills, handler.GetNewBills)
	mux.HandleFunc(tasks.TypeAddBillsPayments, handler.AddBillsAndPayments)
	mux.HandleFunc(tasks.TypeSendBills, handler.SendBills)
	mux.HandleFunc(tasks.TypeGetUnpaidBillIds, handler.GetUnpaidBillIds)
	mux.HandleFunc(tasks.TypeSetDiscordCommands, handler.SetDiscordCommands)
	mux.HandleFunc(tasks.TypeSendNoNewBills, handler.SendNoNewBills)
	mux.HandleFunc(tasks.TypeGetOutstandingBills, handler.GetOutstandingBills)
	mux.HandleFunc(tasks.TypeSendOutstandingBills, handler.SendOutstandingBills)

	// spin up workers
	srv.Run(mux)
}
