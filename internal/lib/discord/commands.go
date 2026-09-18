package discord

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/bwmarrin/discordgo"
)

type (
	InteractionResponse     = discordgo.InteractionResponse
	InteractionResponseData = discordgo.InteractionResponseData
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
					Name:        "bill",
					Description: "The id of bill that was paid",
					Type:        discordgo.ApplicationCommandOptionInteger,
					Required:    true,
					Choices:     billIdChoices,
				},
			},
		},
	}

	if _, err := c.ApplicationCommandBulkOverwrite(c.appId, c.guildId, commands); err != nil {
		fmt.Println(err)
		return fmt.Errorf("Failed to bulk overwrite discord slash commands: %w", err)
	}

	return nil
}

var DiscordToRoomieMap = map[string]string{
	"kanwoody":        "Kane",
	"alexraaee_49709": "Alex",
	"chanceyboyyyyy":  "Chance",
	"madison04547":    "Madison",
}

type Interaction = discordgo.Interaction

type InteractionInfo struct {
	BillId int64
	Roomie string
}

func (c Client) VerifyInteraction(r *http.Request) bool {
	pubKeyBytes, err := hex.DecodeString(c.publicKey)
	if err != nil {
		return false
	}

	return discordgo.VerifyInteraction(r, pubKeyBytes)
}

func (c Client) DecodeInteraction(body io.ReadCloser) (Interaction, error) {
	var payload Interaction
	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return Interaction{}, err
	}

	return payload, nil
}

func (c Client) GetRoomieAndBill(i Interaction, data discordgo.ApplicationCommandInteractionData) (InteractionInfo, error) {
	// Get roomie name from discord name
	discordUser := i.Member.User.Username
	roomie, ok := DiscordToRoomieMap[discordUser]
	if !ok {
		return InteractionInfo{}, fmt.Errorf("Could not map discord user to roomie name: %s", discordUser)
	}

	// Get bill id
	var billId int64
	opts := parseOptions(data.Options)
	if v, ok := opts["bill"]; ok && v.IntValue() != 0 {
		billId = v.IntValue()
	} else {
		return InteractionInfo{}, errors.New("Could not find bill option in interaction data")
	}

	return InteractionInfo{billId, roomie}, nil
}

type optionMap = map[string]*discordgo.ApplicationCommandInteractionDataOption

func parseOptions(options []*discordgo.ApplicationCommandInteractionDataOption) (om optionMap) {
	om = make(optionMap)
	for _, opt := range options {
		om[opt.Name] = opt
	}
	return
}

func (c Client) RespondToDiscordChannel(i *Interaction, res *InteractionResponse) error {
	if err := c.InteractionRespond(i, res); err != nil {
		return fmt.Errorf("Error while responding to interaction: %w", err)
	}

	return nil
}
