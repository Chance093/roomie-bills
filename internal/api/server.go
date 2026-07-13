package api

import (
	"net/http"

	"github.com/Chance093/roomie-bills/internal/db"
)

type Server struct {
	Router *http.ServeMux
	Addr   string
	DB     *db.DB
	env    map[string]string
}

// NewServer intializes a server, sets up routes, and allows database access
// to all handlers associated with that server.
func NewServer(port string, db *db.DB, env map[string]string) *Server {
	// init server
	s := &Server{
		Router: http.NewServeMux(),
		Addr:   ":" + port,
		DB:     db,
		env:    env,
	}

	s.Router.HandleFunc("POST /webhooks/plaid", s.plaidWebhookHandler)
	s.Router.HandleFunc("PUT /bills/{id}", s.plaidWebhookHandler)

	return s
}
