package lib

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

type DiscordClient struct {
	client    *discordgo.Session
	channelId string
}

func NewDiscordClient(env map[string]string) (DiscordClient, error) {
	client, err := discordgo.New("Bot " + env["DISCORD_TOKEN"])
	if err != nil {
		return DiscordClient{}, err
	}
	return DiscordClient{client: client, channelId: env["DISCORD_CHANNEL_ID"]}, nil
}

func (dc *DiscordClient) PostMessage(message string) {
	dc.formatMessage()
	fmt.Println("printing message to discord")
}

func (dc *DiscordClient) formatMessage() {
	fmt.Println("formatting message")
}

func (dc *DiscordClient) SendHostedLink(roomie, hostedLink string) error {
	messageOne := fmt.Sprintf("A link has been requested for %s.\n", roomie)
	messageTwo := fmt.Sprintf("Plaid link: %s", hostedLink)
	finalMessage := messageOne + messageTwo

	if _, err := dc.client.ChannelMessageSend(dc.channelId, finalMessage); err != nil {
		return err
	}

	fmt.Println("Sent hosted link to discord channel")

	return nil
}
