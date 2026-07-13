package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Chance093/roomie-bills/internal/api"
	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/db"
)

const port = "8080"

func main() {
	// init and run server
	db := db.NewDB()
	defer db.Close()

	// get env variables
	env, err := cfg.GetEnv()
	if err != nil {
		log.Fatal(err)
	}

	s := api.NewServer(port, db, env)
	fmt.Printf("Serving on port :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, s.Router))
}
