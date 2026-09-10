package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

type BillPaidPayload struct {
	DiscordRoomie string `json:"discordRoomie"`
}

var DiscordToRoomieMap = map[string]string{
	"kanwoody":      "Kane",
	"Alexraaee":     "Alex",
	"ChanceyBoyyyy": "Chance",
	"Madison":       "Madison",
}

func (s Server) billPaidHandler(w http.ResponseWriter, r *http.Request) {
	billId, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, errors.New("Invalid bill id set in path"))
		return
	}

	// parse request json
	var payload BillPaidPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("error decoding json: %w", err))
		return
	}
	defer r.Body.Close()

	// Get roomie name from discord name
	roomie, ok := DiscordToRoomieMap[payload.DiscordRoomie]
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Errorf("Discord user [%s] not recognized", payload.DiscordRoomie))
		return
	}

	// Mark bill paid by roomie in database
	if err := s.DB.MarkBillPaid(billId, roomie); err != nil {
		writeError(w, http.StatusInternalServerError, errors.New("Internal Server Error"))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
