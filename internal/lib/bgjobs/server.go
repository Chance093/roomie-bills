package bgjobs

type ServerConfig struct {
	Concurrency int
	MaxTimeout  string
}

type Server struct {
	RedisOpts
	cfg ServerConfig
}


// set options to server
func NewServer(opts RedisOpts, cfg ServerConfig) Server {
	return Server{opts, cfg}
}

// continuously tries to pop off queue until task shows up,
// then it will look up the task name in the multiplexer and 
// get the handler to run
func (s *Server) Run(mux ServeMux) {}

// multiplexer that maps task names to task handlers
type ServeMux struct {}

// return ServeMux struct
func NewServeMux() {}

// maps task name to a task handler
func (m *ServeMux) HandleFunc() {}
