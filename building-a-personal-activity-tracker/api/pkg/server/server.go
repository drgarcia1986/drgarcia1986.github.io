package server

import (
	"log"
	"net/http"

	"github.com/drgarcia1986/floki/pkg/influxdb"
)

type Server struct {
	addr   string
	client influxdb.Client
}

func (s *Server) Run() error {
	http.HandleFunc("/track/", s.track)

	log.Printf("Starting server on addr: %s", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

func New(addr string, client influxdb.Client) *Server {
	return &Server{addr: addr, client: client}
}
