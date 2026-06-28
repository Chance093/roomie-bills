package api

import "net/http"

type Server struct {
	Router *http.ServeMux
	Addr   string
	Env    map[string]string
}

// NewServer intializes a server, sets up routes, and allows database access
// to all handlers associated with that server.
func NewServer(port string, env map[string]string) *Server {
	// init server
	s := &Server{
		Router: http.NewServeMux(),
		Addr:   ":" + port,
		Env:    env,
	}

	s.Router.HandleFunc("GET /webhook/plaid", exampleHandler)
	s.Router.HandleFunc("PUT /bills/{id}", exampleHandler)

	return s
}

func exampleHandler(w http.ResponseWriter, r *http.Request) {
	
}

