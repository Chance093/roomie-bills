package api

import (
	"log"
	"net/http"
)

func writeError(w http.ResponseWriter, status int, err error) {
	w.WriteHeader(status)
	w.Write([]byte(err.Error()))
	log.Println(err.Error())
}
