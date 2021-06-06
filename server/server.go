package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/exler/fileigloo/storage"

	"github.com/didip/tollbooth"
	"github.com/gorilla/mux"
)

type OptionFn func(*Server)

func MaxUploadSize(kbytes int64) OptionFn {
	return func(s *Server) {
		s.maxUploadSize = kbytes * 1024
	}
}

func RateLimit(requests int) OptionFn {
	return func(s *Server) {
		s.maxRequests = requests
	}
}

func UseStorage(storage storage.Storage) OptionFn {
	return func(s *Server) {
		s.storage = storage
	}
}

func Port(port int) OptionFn {
	return func(s *Server) {
		s.port = port
	}
}

type Server struct {
	router *mux.Router

	storage storage.Storage

	maxUploadSize int64
	maxRequests   int

	port int
}

func New(options ...OptionFn) *Server {
	s := &Server{}
	for _, optionFn := range options {
		optionFn(s)
	}
	return s
}

func (s *Server) Run() error {
	fs := http.FileServer(http.Dir("./public"))
	limiter := tollbooth.NewLimiter(float64(s.maxRequests), nil)

	s.router = mux.NewRouter()
	s.router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))
	s.router.HandleFunc("/", s.indexHandler).Methods("GET").Name("index")
	s.router.HandleFunc("/", s.uploadHandler).Methods("POST").Name("upload")
	s.router.HandleFunc("/{fileId}", s.downloadHandler).Methods("GET").Name("download")

	log.Println("Server started...")

	err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), tollbooth.LimitHandler(limiter, s.router))
	if err != nil {
		return err
	}

	return nil
}
