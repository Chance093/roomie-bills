package cfg

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func GetEnv() (map[string]string, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("could not load .env file: %w", err)
	}

	envNames := []string{
		"PLAID_CLIENT_ID",
		"PLAID_SANDBOX_SECRET",
		"DISCORD_TOKEN",
		"DISCORD_CHANNEL_ID",
		"DOMAIN",
	}

	env := make(map[string]string)
	for _, name := range envNames {
		val := os.Getenv(name)
		env[name] = val
	}

	return env, nil
}
