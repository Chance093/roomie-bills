package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/lib"
	"github.com/Chance093/roomie-bills/internal/lib/bgjobs"
)

const (
	Addr               = "127.0.0.1:6379"
	TypeGetAccessToken = "get:accessToken"
	TypeGetBank        = "get:bank"
	TypeUpdateBank     = "update:bank"
)

type TaskPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	// worker
	srv := bgjobs.NewServer(bgjobs.RedisOpts{Addr: Addr}, bgjobs.ServerConfig{Concurrency: 5})
	mux := bgjobs.NewServeMux()

	mux.HandleFunc(TypeGetAccessToken, HandleGetAccessTokenTask)
	mux.HandleFunc(TypeGetBank, HandleGetBankTask)
	mux.HandleFunc(TypeUpdateBank, HandleUpdateBankTask)

	srv.Run(mux)
}

type GetAccessTokenPayload struct {
	PublicToken string `json:"publicToken"`
}

func HandleGetAccessTokenTask(ctx context.Context, t bgjobs.Task) error {
	// get payload
	var p GetAccessTokenPayload
	if err := json.Unmarshal(t.Payload, &p); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	// get env variables
	env, err := cfg.GetEnv()
	if err != nil {
		return fmt.Errorf("Could not load cfg: %w", err)
	}

	// get access token from plaid
	pc := lib.NewPlaidClient(env)
	accessToken, err := pc.GetAccessToken(p.PublicToken)
	if err != nil {
		return fmt.Errorf("Could not get access token: %w", err)
	}

	// marshal new payload into json
	jp, err := json.Marshal(GetBankPayload{accessToken})
	if err != nil {
		return fmt.Errorf("Could not marshal struct into json: %w", err)
	}

	// enqueue new task
	newTask := bgjobs.NewTask(TypeGetBank, jp)
	c := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})
	if _, err := c.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	return nil
}

type GetBankPayload struct {
	AccessToken lib.AccessToken `json:"accessToken"`
}

func HandleGetBankTask(ctx context.Context, t bgjobs.Task) error {
	return nil
}

func HandleUpdateBankTask(ctx context.Context, t bgjobs.Task) error {
	return nil
}
