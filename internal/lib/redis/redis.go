package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	rdb *redis.Client
}

type Opts = redis.Options

const (
	defaultAddr     = "127.0.0.1:6379"
	defaultProtocol = 2
)

func NewClient(opts *Opts) *Client {
	// set default opts if empty
	if opts == nil {
		opts = &Opts{
			Addr:     defaultAddr,
			Protocol: defaultProtocol,
		}
	} else {
		if opts.Addr == "" {
			opts.Addr = defaultAddr
		}
		if opts.Protocol == 0 {
			opts.Protocol = defaultProtocol
		}
	}

	rdb := redis.NewClient(opts)

	return &Client{rdb}
}

func (c Client) Close() {
	c.rdb.Close()
}

// LIST COMMANDS

func (c Client) ListPush(ctx context.Context, key string, vals ...any) error {
	return c.rdb.LPush(ctx, key, vals).Err()
}

func (c Client) BlockingListMove(ctx context.Context, src, dest, srcpos, destpos string, timeout time.Duration) (string, error) {
	return c.rdb.BLMove(ctx, src, dest, srcpos, destpos, timeout).Result()
}

func (c Client) ListRemove(ctx context.Context, key string, count int64, val any) error {
	return c.rdb.LRem(ctx, key, count, val).Err()
}

// KV COMMANDS

func (c Client) KVSet(ctx context.Context, key string, val any, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, val, expiration).Err()
}

func (c Client) KVGet(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}
