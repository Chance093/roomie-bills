package bgjobs

/*
- Build out task functions that create new tasks
- Client should be able to enqueue those tasks that the above func built
- Interacts with redis server to push to a queue
*/

type RedisOpts struct {
	addr string
}

type Client struct {
	RedisOpts
}

func NewClient(opts RedisOpts) Client {
	c := Client{opts}

	return c
}

type Task struct {
	endpoint string
	payload  []byte
	retries  int
	timeout  int
}

func (c Client) Enqueue(task Task) error {
	// push task info to redis queue

	return nil
}

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
