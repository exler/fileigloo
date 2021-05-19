package server

import (
	"fmt"
	"log"
	"net/http"
)

type OptionFn func(*Server)

func MaxUploadSize(kbytes int64) OptionFn {
	return func(s *Server) {
		s.maxUploadSize = kbytes * 1024
	}
}

func UploadDirectory(path string) OptionFn {
	return func(s *Server) {
		if path[len(path)-1:] != "/" {
			path += "/"
		}

		s.uploadDirectory = path
	}
}

func Port(port string) OptionFn {
	return func(s *Server) {
		s.port = port
	}
}

type Server struct {
	maxUploadSize   int64
	uploadDirectory string

	port string
}

func New(options ...OptionFn) *Server {
	s := &Server{}

	for _, optionFn := range options {
		optionFn(s)
	}

	return s
}

func (s *Server) Run() error {
	log.Println("Server started...")

	http.HandleFunc("/", s.indexHandler)
	http.HandleFunc("/upload", s.uploadHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.port), nil)
	if err != nil {
		return err
	}

	return nil
}
