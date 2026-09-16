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
		b, err := json.Marshal(discord.InteractionResponse{
			Type: discord.InteractionResponsePong,
		})
		if err != nil {
			fmt.Println(err)
			writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		w.Write(b)
		return
	}

	// slash command interaction
	if interaction.Type == discord.InteractionApplicationCommand {
		info, err := s.dc.GetRoomieAndBill(interaction)
		if err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}

		// Mark bill paid by roomie in database
		if err := s.DB.MarkBillPaid(info.BillId, info.Roomie); err != nil {
			writeError(w, http.StatusInternalServerError, errors.New("Internal Server Error"))
			return
		}

		// get unpaid bill id's and reset /paid command
		unpaidBillIds, err := s.DB.GetUnpaidBillIds()
		if err != nil {
			writeError(w, http.StatusInternalServerError, errors.New("Internal Server Error"))
			return
		}

		if err := s.dc.SetCommands(unpaidBillIds); err != nil {
			writeError(w, http.StatusBadGateway, fmt.Errorf("Bad Gateway: %w", err))
			return
		}

		// respond to discord channel
		if err := s.dc.RespondToDiscordChannel(&interaction, &discord.InteractionResponse{
			Type: discord.InteractionResponseChannelMessageWithSource,
			Data: &discord.InteractionResponseData{
				Content: "Thank you for making a payment :)",
			},
		}); err != nil {
			writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
			return
		}

		return
	}

	// send error for unknown interaction
	writeError(w, http.StatusBadRequest, errors.New("Unknown interaction type"))
	return
}
