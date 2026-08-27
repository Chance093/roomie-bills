package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Chance093/roomie-bills/internal/tasks"
)

const Addr = "127.0.0.1:6379"

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
	// get payload and validate
	raw, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("error reading body: %w", err))
		return
	}
	defer r.Body.Close()

	var notif WebhookNotif
	if err := json.Unmarshal(raw, &notif); err != nil {
		writeError(w, http.StatusUnprocessableEntity, fmt.Errorf("error unmarshaling json: %w", err))
		return
	}

	// early return 200 if its not the event we are looking for
	if notif.WebhookType != "LINK" || notif.WebhookCode != "SESSION_FINISHED" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// validate jwt, ip, and payload hash
	ip := r.RemoteAddr
	if strings.Contains(ip, "[::1]") { // ngrok
		ip = getHeaderCI(r.Header, "X-Forwarded-For")
	}
	if err := verifyWebhook(raw, ip, r.Header, s.pc); err != nil {
		writeError(w, http.StatusForbidden, fmt.Errorf("error verifying webhook: %w", err))
		return
	}

	// validate and obtain public token from payload
	if len(notif.PublicTokens) == 0 || notif.PublicTokens[0] == "" {
		writeError(w, http.StatusBadRequest, errors.New("no public token found"))
		return
	}
	publicToken := notif.PublicTokens[0]

	// create new background task and enqueue
	newTask, err := tasks.NewGetAccessTokenTask(publicToken, notif.LinkToken)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	if _, err := s.jc.Enqueue(newTask); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("Could not enqueue new task: %w", err))
		return
	}

	// send back 200
	w.WriteHeader(http.StatusOK)
}
