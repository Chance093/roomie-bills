package api

import (
	"errors"
	"fmt"
	"net/http"
)

func (s Server) billPaidHandler(w http.ResponseWriter, r *http.Request) {
	info, err := s.dc.GetRoomieAndBill(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("Error while getting roomie and bill id: %w", err))
		return
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
