package api

import (
	"net/http"

	"github.com/Chance093/roomie-bills/internal/db"
)

type Server struct {
	Router *http.ServeMux
	Addr   string
	DB     *db.DB
}

// NewServer intializes a server, sets up routes, and allows database access
// to all handlers associated with that server.
func NewServer(port string, db *db.DB) *Server {
	// init server
	s := &Server{
		Router: http.NewServeMux(),
		Addr:   ":" + port,
		DB: db,
	}

	s.Router.HandleFunc("POST /webhooks/plaid", plaidWebhookHandler)
	s.Router.HandleFunc("PUT /bills/{id}", exampleHandler)

	return s
}

func exampleHandler(w http.ResponseWriter, r *http.Request) {
	
}

