package server

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
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

func Purge(older, interval int) OptionFn {
	return func(s *Server) {
		s.purgeOlder = time.Duration(older) * time.Hour
		s.purgeInterval = interval
	}
}

func UseStorage(storage Storage) OptionFn {
	return func(s *Server) {
		s.storage = storage
	}
}

func Port(port string) OptionFn {
	return func(s *Server) {
		s.port = port
	}
}

type Server struct {
	scheduler *gocron.Scheduler
	storage   Storage

	maxUploadSize int64
	maxRequests   int

	purgeOlder    time.Duration
	purgeInterval int

	visitors      map[string]*Visitor
	visitorsMutex sync.Mutex

	port string
}

func New(options ...OptionFn) *Server {
	s := &Server{}

	for _, optionFn := range options {
		optionFn(s)
	}

	s.visitors = make(map[string]*Visitor)

	s.scheduler = gocron.NewScheduler(time.UTC)
	if s.maxRequests != 0 {
		s.scheduler.Every(1).Minute().Do(s.cleanVisitors)
	}
	if s.purgeInterval != 0 {
		s.scheduler.Every(s.purgeInterval).Hours().Do(s.storage.Purge, s.purgeOlder)
	}

	return s
}

func (s *Server) Run() error {
	router := mux.NewRouter()
	router.HandleFunc("/", s.indexHandler).Methods("GET")
	router.HandleFunc("/upload", s.uploadHandler).Methods("POST")
	router.HandleFunc("/{fileId}", s.downloadHandler).Methods("GET")
	router.Use(s.limitMiddleware)
	http.Handle("/", router)

	log.Println("Server started...")

	s.scheduler.StartAsync()

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.port), nil)
	if err != nil {
		return err
	}

	return nil
}
