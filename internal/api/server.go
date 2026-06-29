package api

import (
	"net/http"

	"github.com/Chance093/roomie-bills/internal/db"
)

type Server struct {
	Router *http.ServeMux
	Addr   string
	Env    map[string]string
	DB     *db.DB
}

// NewServer intializes a server, sets up routes, and allows database access
// to all handlers associated with that server.
func NewServer(port string, db *db.DB, env map[string]string) *Server {
	// init server
	s := &Server{
		Router: http.NewServeMux(),
		Addr:   ":" + port,
		Env:    env,
		DB: db,
	}

	s.Router.HandleFunc("GET /webhook/plaid", exampleHandler)
	s.Router.HandleFunc("PUT /bills/{id}", exampleHandler)

	return s
}

func exampleHandler(w http.ResponseWriter, r *http.Request) {
	
}

