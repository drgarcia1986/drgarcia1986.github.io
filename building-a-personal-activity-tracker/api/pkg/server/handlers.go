package server

import (
	"encoding/json"
	"log"
	"net/http"
)

func (s *Server) track(w http.ResponseWriter, r *http.Request) {
	var d requestData
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		log.Printf("Error decoding data: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := s.client.Write(d.Class, d.Title); err != nil {
		log.Printf("Error send data to influxdb: %v", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
