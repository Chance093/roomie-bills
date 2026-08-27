package api

import (
	"encoding/json"
	"io"
	"log"
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
		w.WriteHeader(http.StatusUnprocessableEntity)
		log.Printf("error reading body: %s\n", err.Error())
		return
	}
	defer r.Body.Close()

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

	// validate jwt, ip, and payload hash
	ip := r.RemoteAddr
	if strings.Contains(ip, "[::1]") { // ngrok
		ip = getHeaderCI(r.Header, "X-Forwarded-For")
	}
	if err := verifyWebhook(raw, ip, r.Header, s.pc); err != nil {
		w.WriteHeader(http.StatusForbidden)
		log.Printf("error verifying webhook: %s\n", err.Error())
		return
	}

	// validate and obtain public token from payload
	if len(notif.PublicTokens) == 0 || notif.PublicTokens[0] == "" {
		w.WriteHeader(http.StatusBadRequest)
		log.Println("no public token found")
		return
	}
	publicToken := notif.PublicTokens[0]

	// create new background task and enqueue
	newTask, err := tasks.NewGetAccessTokenTask(publicToken, notif.LinkToken)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	if _, err := s.jc.Enqueue(newTask); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("Could not enqueue new task: %s\n", err.Error())
		return
	}

	// send back 200
	w.WriteHeader(http.StatusOK)
}
