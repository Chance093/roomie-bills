package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Chance093/roomie-bills/internal/api"
	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/discord"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

const port = "8080"

func main() {
	// init everything
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatalf("Could not get env variables: %s\n", err.Error())
	}
	pc := plaid.NewClient(env)
	jc := bgjobs.NewClient(bgjobs.RedisOpts{})
	defer jc.Close()
	db := db.NewDB()
	defer db.Close()

	dc, err := discord.NewClient(env)
	if err != nil {
		log.Fatalf("Could not connect to discord client: %s\n", err.Error())
	}

	// run server
	s := api.NewServer(port, pc, jc, dc, db)
	fmt.Printf("Serving on port :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, s.Router))
}
