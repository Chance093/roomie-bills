package discord

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"

	"github.com/bwmarrin/discordgo"
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

type InteractionInfo struct {
	BillId int64
	Roomie string
}

func (c Client) GetRoomieAndBill(body io.ReadCloser) (InteractionInfo, error) {
	var payload discordgo.Interaction
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return InteractionInfo{}, nil
	}

	// Get roomie name from discord name
	discordUser := payload.Member.User.Username
	roomie, ok := DiscordToRoomieMap[discordUser]
	if !ok {
		return InteractionInfo{}, nil
	}

	// TODO: get actual value
	return InteractionInfo{0, roomie}, nil
}
