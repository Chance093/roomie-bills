package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Chance093/roomie-bills/internal/api"
	"github.com/Chance093/roomie-bills/internal/db"
)

const port = "8080"

func main() {
	// init and run server
	db := db.NewDB()
	defer db.Close()

	s := api.NewServer(port, db)
	fmt.Printf("Serving on port :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, s.Router))
}
