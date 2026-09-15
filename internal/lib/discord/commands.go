package discord

import (
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

var (
	InteractionPing                             = discordgo.InteractionPing
	InteractionApplicationCommand               = discordgo.InteractionApplicationCommand
	InteractionResponsePong                     = discordgo.InteractionResponsePong
	InteractionResponseChannelMessageWithSource = discordgo.InteractionResponseChannelMessageWithSource
)

func (c Client) SetCommands(billIds []int64) error {
	billIdChoices := make([]*discordgo.ApplicationCommandOptionChoice, len(billIds))
	for i, billId := range billIds {
		billIdChoices[i] = &discordgo.ApplicationCommandOptionChoice{
			Name:  strconv.Itoa(int(billId)), // convert int64 to string
			Value: billId,
		}
	}

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "paid",
			Description: "Mark a bill paid by roomie",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "Bill Id",
					Description: "The id of bill that was paid",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
					Choices:     billIdChoices,
				},
			},
		},
	}

	// TODO: fill in appId and guildId
	if _, err := c.ApplicationCommandBulkOverwrite("", "", commands); err != nil {
		return fmt.Errorf("Failed to bulk overwrite discord slash commands: %w", err)
	}

	return nil
}

var DiscordToRoomieMap = map[string]string{
	"kanwoody":      "Kane",
	"Alexraaee":     "Alex",
	"ChanceyBoyyyy": "Chance",
	"Madison":       "Madison",
}

type Interaction = discordgo.Interaction

type InteractionInfo struct {
	BillId int64
	Roomie string
}

func (c Client) VerifyInteraction(r *http.Request) bool {
	return discordgo.VerifyInteraction(r, ed25519.PublicKey("")) // TODO: fill in with public key
}

func (c Client) DecodeInteraction(body io.ReadCloser) (Interaction, error) {
	var payload Interaction
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return Interaction{}, err
	}

	return payload, nil
}

func (c Client) GetRoomieAndBill(payload Interaction) (InteractionInfo, error) {
	// Get roomie name from discord name
	discordUser := payload.Member.User.Username
	roomie, ok := DiscordToRoomieMap[discordUser]
	if !ok {
		return InteractionInfo{}, nil
	}

	// TODO: get actual value
	return InteractionInfo{0, roomie}, nil
}
