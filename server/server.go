package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/exler/fileigloo/storage"
	"golang.org/x/crypto/acme/autocert"

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

func HTTPS(domains []string) OptionFn {
	return func(s *Server) {
		log.Println("HTTPS: enabled")

		hostPolicy := func(ctx context.Context, host string) error {
			for _, d := range domains {
				if strings.HasSuffix(host, d) {
					return nil
				}
			}

			return errors.New("acme/autocert: hosts not allowed")
		}

		m := &autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: hostPolicy,
			Cache:      autocert.DirCache("./cache"),
		}

		s.tlsConfig = &tls.Config{GetCertificate: m.GetCertificate}
	}
}

type Server struct {
	tlsConfig *tls.Config

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

	log.Printf("Storage type: %s\n", s.storage.Type())

	return s
}

func (s *Server) Run() {
	fs := http.FileServer(http.Dir("./public"))
	limiter := tollbooth.NewLimiter(float64(s.maxRequests), nil)

	s.router = mux.NewRouter()
	s.router.PathPrefix("/public/").Handler(http.StripPrefix("/public/", fs))
	s.router.HandleFunc("/", s.indexHandler).Methods("GET").Name("index")
	s.router.HandleFunc("/", s.uploadHandler).Methods("POST").Name("upload")
	s.router.HandleFunc("/{raw:(?:raw)}/{fileId}", s.downloadHandler).Methods("GET").Name("download-raw")
	s.router.HandleFunc("/{fileId}", s.downloadHandler).Methods("GET").Name("download")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      tollbooth.LimitHandler(limiter, s.router),
		TLSConfig:    s.tlsConfig,
	}
	log.Println("http: Server started")

	if s.tlsConfig != nil {
		go func() {
			err := srv.ListenAndServeTLS("", "")
			if err != nil {
				log.Fatalln(err.Error())
			}
		}()
	} else {
		go func() {
			err := srv.ListenAndServe()
			if err != nil {
				log.Fatalln(err.Error())
			}
		}()
	}

	c := make(chan os.Signal, 1)
	// Accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	signal.Notify(c, os.Interrupt)

	// Block until signal received
	<-c

	// Wait 10 second for existing connections to finish
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	// Does not block if no connections, otherwises waits for timeout
	srv.Shutdown(ctx)
}
