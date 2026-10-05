package api

import (
	"net/http"

	"github.com/Chance093/roomie-bills/internal/db"
	"github.com/Chance093/roomie-bills/internal/lib/discord"
	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/lib/taskqueue"
)

type Server struct {
	Router *http.ServeMux
	Addr   string
	DB     *db.DB
	pc     plaid.Client
	tc     taskqueue.Client
	dc     discord.Client
}

// NewServer intializes a server, sets up routes, and allows database access
// to all handlers associated with that server.
func NewServer(port string, pc plaid.Client, tc taskqueue.Client, dc discord.Client, db *db.DB) *Server {
	// init server
	s := &Server{
		Router: http.NewServeMux(),
		Addr:   ":" + port,
		DB:     db,
		pc:     pc,
		tc:     tc,
		dc:     dc,
	}

	s.Router.HandleFunc("POST /webhooks/plaid", s.plaidWebhookHandler)
	s.Router.HandleFunc("POST /discord/interactions", s.billPaidHandler)

	return s
}
