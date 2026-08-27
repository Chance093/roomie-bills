package main

import (
	"log"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/tasks"
)

const Addr = "127.0.0.1:6379"

func main() {
	// init everything
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatalf("Could not get env variables: %s\n", err.Error())
	}
	pc := lib.NewPlaidClient(env)
	jc := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})
	db := db.NewDB()
	defer db.Close()

	// config server and handlers
	srv := bgjobs.NewServer(bgjobs.RedisOpts{Addr: Addr}, bgjobs.ServerConfig{Concurrency: 5})
	mux := bgjobs.NewServeMux()
	handler := tasks.NewHandler(pc, jc, db)

	mux.HandleFunc(tasks.TypeGetAccessToken, handler.GetAccessToken)
	mux.HandleFunc(tasks.TypeGetBank, handler.GetBankName)
	mux.HandleFunc(tasks.TypeUpdateBank, handler.UpdateBank)

	// spin up workers
	srv.Run(mux)
}
