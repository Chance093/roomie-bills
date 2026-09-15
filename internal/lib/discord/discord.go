package discord

import "github.com/bwmarrin/discordgo"

type Client struct {
	*discordgo.Session
	appId     string
	publicKey string
	channelId string
	guildId   string
}

func NewClient(env map[string]string) (Client, error) {
	client, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	if err != nil {
		return Client{}, err
	}

	return Client{
		Session:   client,
		appId:     env["DISCORD_APP_ID"],
		publicKey: env["DISCORD_PUBLIC_KEY"],
		channelId: env["DISCORD_CHANNEL_ID"],
		guildId:   env["DISCORD_GUILD_ID"],
	}, nil
}
