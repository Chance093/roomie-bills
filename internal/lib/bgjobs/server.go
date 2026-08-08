package bgjobs

/*
- Spins up worker processes
- Listens to redis to see if it can pull anything
- If it pulls something, put task in temp queue until task has been confirmed completed
- If task not confirmed completed within amount of time, push back to og queue
- If task fails a certain amount of times, push to DLQ
- Add retry logic
*/

type ServerConfig struct {
	Concurrency int
	MaxTimeout  string
}

type Server struct {
	RedisOpts
	cfg ServerConfig
}

func NewServer(opts RedisOpts, cfg ServerConfig) Server {
	return Server{opts, cfg}
}
