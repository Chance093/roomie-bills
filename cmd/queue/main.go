package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Chance093/roomie-bills/internal/cfg"
	"github.com/Chance093/roomie-bills/internal/db"
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
	LinkToken   string `json:"linkToken"`
}

func HandleGetAccessTokenTask(ctx context.Context, t bgjobs.Task) error {
	fmt.Println("Starting HandleGetAccessTokenTask")
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
	jp, err := json.Marshal(GetBankPayload{accessToken, p.LinkToken})
	if err != nil {
		return fmt.Errorf("Could not marshal struct into json: %w", err)
	}

	fmt.Println("Enqueueing GetBankTask")
	// enqueue new task
	newTask := bgjobs.NewTask(TypeGetBank, jp)
	c := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})
	if _, err := c.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	fmt.Println("Finished HandleGetAccessTokenTask")
	return nil
}

type GetBankPayload struct {
	AccessToken lib.AccessToken `json:"accessToken"`
	LinkToken   string          `json:"linkToken"`
}

func HandleGetBankTask(ctx context.Context, t bgjobs.Task) error {
	fmt.Println("Starting HandleGetBankTask")
	// get payload
	var p GetBankPayload
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
	bank, err := pc.GetBankName(p.AccessToken)
	if err != nil {
		return fmt.Errorf("Could not get bank name from plaid: %w", err)
	}

	// marshal new payload into json
	jp, err := json.Marshal(UpdateBankRecordPayload{
		LinkToken:   p.LinkToken,
		AccessToken: p.AccessToken.Token,
		ItemId:      p.AccessToken.ItemId,
		Bank:        bank,
	})
	if err != nil {
		return fmt.Errorf("Could not marshal struct into json: %w", err)
	}

	// enqueue new task
	fmt.Println("Enqueueing UpdateBankTask")
	newTask := bgjobs.NewTask(TypeUpdateBank, jp)
	c := bgjobs.NewClient(bgjobs.RedisOpts{Addr: Addr})
	if _, err := c.Enqueue(newTask); err != nil {
		return fmt.Errorf("Could not enqueue new task: %w", err)
	}

	fmt.Println("Finished HandleGetBankTask")
	return nil
}

type UpdateBankRecordPayload struct {
	LinkToken   string `json:"linkToken"`
	AccessToken string `json:"accessToken"`
	ItemId      string `json:"itemId"`
	Bank        string `json:"bank"`
}

func HandleUpdateBankTask(ctx context.Context, t bgjobs.Task) error {
	fmt.Println("Starting HandleUpdateBankTask")
	// get payload
	var p UpdateBankRecordPayload
	if err := json.Unmarshal(t.Payload, &p); err != nil {
		return fmt.Errorf("Could not unmarshal json into payload: %w", err)
	}

	db := db.NewDB()
	defer db.Close()

	// check if bank record has already been updated (idempotent)
	if exists, err := db.AccessTokenExists(p.AccessToken); err != nil {
		return fmt.Errorf("Error while checking if access token already exists in db: %w", err)
	} else if exists {
		return nil // access token exists and bank record has already been updated
	}

	if err := db.UpdateBankRecord(p.LinkToken, p.AccessToken, p.ItemId, p.Bank); err != nil {
		return fmt.Errorf("Could not update bank record: %w", err)
	}

	fmt.Println("Finished HandleUpdateBankTask")
	return nil
}
