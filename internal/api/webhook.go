package api

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/Chance093/roomie-bills/internal/lib"
)

type WebhookNotif struct {
	WebhookType   string   `json:"webhook_type"`
	WebhookCode   string   `json:"webhook_code"`
	Status        string   `json:"status"`
	LinkSessionId string   `json:"link_session_id"`
	LinkToken     string   `json:"link_token"`
	PublicTokens  []string `json:"public_tokens"`
	Environment   string   `json:"environment"`
}

func (s *Server) plaidWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// get payload, header, and ip for validation
	ip := r.RemoteAddr
	if strings.Contains(ip, "[::1]") { // ngrok
		ip = getHeaderCI(r.Header, "X-Forwarded-For")
	}
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Printf("error reading body: %s\n", err.Error())
		return
	}
	defer r.Body.Close()

	// validate jwt, ip, and payload hash
	pc := lib.NewPlaidClient(s.env)
	ok, err := verifyWebhook(raw, ip, r.Header, pc)
	// WARN: SHOULD THESE BE IN SAME IF STATEMENT?
	if err != nil || !ok {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("error verifying webhook: %s\n", err.Error())
		return
	}

	// validate expected payload
	var notif WebhookNotif
	if err := json.Unmarshal(raw, &notif); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Printf("error unmarshaling json: %s\n", err.Error())
		return
	}

	// early return 200 if its not the event we are looking for
	if notif.WebhookType != "LINK" || notif.WebhookCode != "SESSION_FINISHED" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// validate and obtain public token from payload
	// WARN: This might crash program because of 2nd conditional (check online)
	if len(notif.PublicTokens) == 0 || notif.PublicTokens[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("no public token found")
		return
	} 
	publicToken := notif.PublicTokens[0]

	// TODO: turn everything below into a background task

	// get access token, bank name and save to db
	accessToken, err := pc.GetAccessToken(publicToken)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		log.Printf("error getting access token: %s", err.Error())
		return
	}

	bank, err := pc.GetBankName(accessToken)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		log.Printf("error getting bank name: %s", err.Error())
		return
	}

	if err = s.DB.UpdateBankRecord(notif.LinkToken, accessToken.Token, accessToken.ItemId, bank); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error updating bank record: %s", err.Error())
		return
	}

	// send back 200
	w.WriteHeader(http.StatusOK)
}
