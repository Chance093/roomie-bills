package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/Chance093/roomie-bills/internal/lib/discord"
)

func (s Server) billPaidHandler(w http.ResponseWriter, r *http.Request) {
	// verify headers and decode interaction payload
	if ok := s.dc.VerifyInteraction(r); !ok {
		writeError(w, http.StatusUnauthorized, errors.New("Unauthorized"))
		return
	}

	interaction, err := s.dc.DecodeInteraction(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("Error while getting roomie and bill id: %w", err))
		return
	}
	defer r.Body.Close()

	// ping interaction
	if interaction.Type == discord.InteractionPing {
		type PongResponse struct {
			Type discord.InteractionResponseType `json:"type"`
		}

		b, err := json.Marshal(PongResponse{discord.InteractionResponsePong})
		if err != nil {
			writeError(w, http.StatusInternalServerError, errors.New("Internal Server Error"))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(b)
		return
	}

	// slash command interaction
	if interaction.Type == discord.InteractionApplicationCommand {
		info, err := s.dc.GetRoomieAndBill(interaction)
		if err != nil {
		}

		// Mark bill paid by roomie in database
		if err := s.DB.MarkBillPaid(info.BillId, info.Roomie); err != nil {
			writeError(w, http.StatusInternalServerError, errors.New("Internal Server Error"))
			return
		}

		// TODO: do actual db search to get these
		unpaidBillIds := []int64{}

		if err := s.dc.SetCommands(unpaidBillIds); err != nil {
			writeError(w, http.StatusBadGateway, fmt.Errorf("Bad Gateway: %w", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
		return
	}

	// TODO: error handling here (unknown interaction)
}
