package main

import (
	"log"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/tasks"
)

func main() {
	// init everything
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatalf("Could not get env variables: %s\n", err.Error())
	}

	pc := plaid.NewClient(env)
  dc, err := lib.NewDiscordClient(env)
  if err != nil {
		log.Fatalf("Could not connect to discord client: %s\n", err.Error())
  }

	redisOpts := bgjobs.RedisOpts{}
	jc := bgjobs.NewClient(redisOpts)
	defer jc.Close()

	db := db.NewDB()
	defer db.Close()

	// config server and handlers
	srv := bgjobs.NewServer(redisOpts, bgjobs.ServerConfig{})
	mux := bgjobs.NewServeMux()
	handler := tasks.NewHandler(pc, jc, dc, db)

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
	mux.HandleFunc(tasks.TypeSendNoBills, handler.SendNoBills)

	// spin up workers
	srv.Run(mux)
}
