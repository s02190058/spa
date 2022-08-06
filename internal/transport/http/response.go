package http

import (
	"encoding/json"
	"log"
	"net/http"
)

func errorResponse(w http.ResponseWriter, code int, err error) {
	response(w, code, map[string]string{
		"message": err.Error(),
	})
}

func response(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// TODO: change the default logger
		log.Printf("json.NewEncoder: %v", err)
	}
}
