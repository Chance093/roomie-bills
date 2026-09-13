package discord

import "github.com/bwmarrin/discordgo"

type Client struct {
	*discordgo.Session
	channelId string
}

func NewClient(env map[string]string) (Client, error) {
	client, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	if err != nil {
		return Client{}, err
	}

	return Client{client, env["DISCORD_CHANNEL_ID"]}, nil
}
