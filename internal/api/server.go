package api

import (
	"net/http"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
)

type Server struct {
	Router *http.ServeMux
	Addr   string
	DB     *db.DB
	pc     plaid.Client
	jc     bgjobs.Client
}

// NewServer intializes a server, sets up routes, and allows database access
// to all handlers associated with that server.
func NewServer(port string, pc plaid.Client, jc bgjobs.Client, db *db.DB) *Server {
	// init server
	s := &Server{
		Router: http.NewServeMux(),
		Addr:   ":" + port,
		DB:     db,
		pc:     pc,
		jc:     jc,
	}

	s.Router.HandleFunc("POST /webhooks/plaid", s.plaidWebhookHandler)
	s.Router.HandleFunc("PUT /bills/{id}", s.plaidWebhookHandler)

	return s
}
