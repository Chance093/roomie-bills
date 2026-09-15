package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Chance093/roomie-bills/internal/lib/discord"
)

func (s Server) billPaidHandler(w http.ResponseWriter, r *http.Request) {
	if ok := s.dc.VerifyInteraction(r); !ok {
		// TODO: error handling
		return
	}

	payload, err := s.dc.DecodeInteraction(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("Error while getting roomie and bill id: %w", err))
		return
	}

	// TODO: ping request, send back pong
	if payload.Type == discord.InteractionPing {
	}

	if payload.Type == discord.InteractionApplicationCommand {
		info, err := s.dc.GetRoomieAndBill(payload)
		if err != nil {
		}
		defer r.Body.Close()

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
	}

	// TODO: error handling here (unknown interaction)
}
