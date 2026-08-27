package bgjobs

import "github.com/redis/go-redis/v9"

const (
	DefaultAddr = "127.0.0.1:6379"
	Primary     = "jobs:queue"
	Temp        = "jobs:temp"
	DLQ         = "jobs:dlq"
)

type RedisOpts struct {
	Addr     string
	Password string
	DB       int
	Protocol int
}

func initNewRDBClient(opts RedisOpts) *redis.Client {
	// set default options
	if opts.Addr == "" {
		opts.Addr = DefaultAddr
	}
	if opts.Protocol == 0 {
		opts.Protocol = 2
	}

	return redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
		Protocol: opts.Protocol,
	})
}
