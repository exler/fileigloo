package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/exler/fileigloo/logger"
	"github.com/exler/fileigloo/storage"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type OptionFn func(*Server)

func MaxUploadSize(kbytes int64) OptionFn {
	return func(s *Server) {
		s.maxUploadSize = kbytes * 1024
	}
}

func MaxRequests(requests int) OptionFn {
	return func(s *Server) {
		s.maxRequests = requests
	}
}

func UseStorage(storage storage.Storage) OptionFn {
	return func(s *Server) {
		s.storage = storage
	}
}

func UseLogger(logger *logger.Logger) OptionFn {
	return func(s *Server) {
		s.logger = logger
	}
}

func Port(port int) OptionFn {
	return func(s *Server) {
		s.port = port
	}
}

type Server struct {
	logger *logger.Logger

	router chi.Router

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

func (s *Server) Run() {
	fs := http.FileServer(http.Dir("./public"))

	s.router = chi.NewRouter()
	s.router.Use(httprate.LimitByIP(s.maxRequests, 1*time.Minute))

	s.router.Handle("/public/*", http.StripPrefix("/public/", fs))
	s.router.Get("/", s.indexHandler)
	s.router.Post("/upload", s.uploadHandler)
	s.router.Post("/pastebin", s.pastebinHandler)
	s.router.Get("/{view:(?:view)}/{fileId}", s.downloadHandler)
	s.router.Get("/{fileId}", s.downloadHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router,
	}
	s.logger.Debug(fmt.Sprintf("Server started [storage=%s]", s.storage.Type()))

	err := s.storage.Purge()
	if err != nil {
		s.logger.Error(err)
	}

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			s.logger.Error(err)
			return
		}
	}()

	c := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(c, os.Interrupt)

	// Block until signal received
	<-c

	// Wait 10 second for existing connections to finish
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Does not block if no connections, otherwises waits for timeout
	srv.Shutdown(ctx) //#nosec
}
