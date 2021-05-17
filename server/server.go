package server

import (
	"fmt"
	"log"
	"net/http"
)

type Server struct {
	maxUploadSize int64

	port string
}

func New() *Server {
	s := &Server{}
	return s
}

func (s *Server) Run() error {
	log.Println("Server started...")

	http.HandleFunc("/", s.indexHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.port), nil)
	if err != nil {
		return err
	}

	return nil
}
