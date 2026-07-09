package api

import (
	"encoding/json"
	"net/http"
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

var validIPs = [4]string{"52.21.26.131", "52.21.47.157", "52.41.247.19", "52.88.82.239"}

func plaidWebhookHandler(w http.ResponseWriter, r *http.Request) {
	// validate ip
	ip := r.RemoteAddr
	if ip != validIPs[0] && ip != validIPs[1] && ip != validIPs[2] && ip != validIPs[3] {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// verify jwt in header

	// validate expected payload
	var notif WebhookNotif
	if err := json.NewDecoder(r.Body).Decode(&notif); err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	defer r.Body.Close()

	// grab public token from payload
	publicToken := notif.PublicTokens[0]
	if publicToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// start background task to get access token

	// send back 200
	w.WriteHeader(http.StatusOK)
}

// background task
func getAccessToken() {
	// take public token
	// hit /item/public_token/exchange
	// get access token and save to db
	// send discord notif that bank account has been connected
}
