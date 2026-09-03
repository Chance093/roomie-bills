package lib

import (
	"fmt"
	"strings"

	"github.com/Chance093/roomie-bills/internal/lib/plaid"
	"github.com/Chance093/roomie-bills/internal/types"
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

type SplitBill struct {
	plaid.Bill
	Split float64
}

func (dc *DiscordClient) SendBills(bills []types.SplitBill) error {
	var b strings.Builder
	b.WriteString("```")
	b.WriteString("New Bills:\n\n")

	for _, bill := range bills {
		// TODO: fix help messages
		b.WriteString(fmt.Sprintf("#️⃣ Bill ID: %s\n", bill.Id))
		b.WriteString(fmt.Sprintf("📋 New Bill: %s\n", bill.Payee))
		b.WriteString(fmt.Sprintf("📅 Date: %s\n", bill.Date))
		b.WriteString(fmt.Sprintf("💰 Total: $%.2f\n", bill.Total))
		b.WriteString(fmt.Sprintf("👤 Each roommate owes Chance: $%.2f\n", bill.Split))
		b.WriteString("\n")
	}

	b.WriteString("React with ✅ when you've paid Chance back!\n")
	b.WriteString("— Kane | Alex | Madison")
	b.WriteString("```")

	if _, err := dc.client.ChannelMessageSend(dc.channelId, b.String()); err != nil {
		return err
	}

	return nil
}

func (dc *DiscordClient) SendNoBillsMessage() error {
	message := "```All previous bills caught up :)```"

	if _, err := dc.client.ChannelMessageSend(dc.channelId, message); err != nil {
		return err
	}

	return nil
}
