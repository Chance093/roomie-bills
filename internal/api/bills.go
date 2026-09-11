package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type DiscordInteractionPayload struct {
	Type   int                    `json:"type"`
	Data   DiscordInteractionData `json:"data"`
	Member DiscordGuildMember     `json:"member"`
}

type DiscordInteractionData struct {
	Name    string                `json:"name"`
	Type    int                   `json:"type"`
	Options DiscordCommandOptions `json:"options"`
}

type DiscordCommandOptions struct {
	Name  string `json:"name"`
	Type  int    `json:"type"`
	Value int64  `json:"value"`
}

type DiscordGuildMember struct {
	User DiscordUser `json:"user"`
}

type DiscordUser struct {
	Username string `json:"username"`
}

var DiscordToRoomieMap = map[string]string{
	"kanwoody":      "Kane",
	"Alexraaee":     "Alex",
	"ChanceyBoyyyy": "Chance",
	"Madison":       "Madison",
}

func (s Server) billPaidHandler(w http.ResponseWriter, r *http.Request) {
  // do some validation here
	// parse request json
	var payload DiscordInteractionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("error decoding json: %w", err))
		return
	}
	defer r.Body.Close()

	// Get roomie name from discord name
  discordUser := payload.Member.User.Username
	roomie, ok := DiscordToRoomieMap[discordUser]
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Errorf("Discord user [%s] not recognized", discordUser))
		return
	}

	// Mark bill paid by roomie in database
	if err := s.DB.MarkBillPaid(payload.Data.Options.Value, roomie); err != nil {
		writeError(w, http.StatusInternalServerError, errors.New("Internal Server Error"))
		return
	}

	// TODO: update bills command in discord to show only unpaid bills

	w.WriteHeader(http.StatusNoContent)
}
