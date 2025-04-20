package server

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/exler/fileigloo/logger"
	"github.com/exler/fileigloo/storage"
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"golang.org/x/crypto/bcrypt"
)

//go:embed static/*
var StaticFS embed.FS

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

func Sentry(sentryDSN, sentryEnvironment string, sentryTracesSampleRate float64) OptionFn {
	return func(s *Server) {
		if sentryDSN == "" {
			return
		}

		err := sentry.Init(sentry.ClientOptions{
			Dsn:              sentryDSN,
			Environment:      sentryEnvironment,
			EnableTracing:    sentryTracesSampleRate > 0,
			TracesSampleRate: sentryTracesSampleRate,
		})
		if err != nil {
			s.logger.Error(err)
		}
		defer sentry.Flush(time.Second)
	}
}

func Port(port int) OptionFn {
	return func(s *Server) {
		s.port = port
	}
}

func SitePassword(password string) OptionFn {
	return func(s *Server) {
		if password == "" {
			return
		}

		sitePasswordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			s.logger.Error(err)
		}

		s.sitePasswordHash = string(sitePasswordHash)
	}
}

type Server struct {
	logger *logger.Logger

	router chi.Router

	// protectedRouter is not necessarily protected, only if the SitePassword option is used
	protectedRouter chi.Router

	storage storage.Storage

	maxUploadSize int64
	maxRequests   int

	sitePasswordHash string

	port int
}

func New(options ...OptionFn) *Server {
	s := &Server{
		logger: logger.NewLogger(),
	}
	for _, optionFn := range options {
		optionFn(s)
	}
	return s
}

func (s *Server) Run() {
	fs := http.FileServer(http.FS(StaticFS))

	limiter := httprate.LimitByIP(s.maxRequests, 1*time.Minute)
	sentryMiddleware := sentryhttp.New(sentryhttp.Options{
		Repanic: true,
	})

	s.router = chi.NewRouter()
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(limiter)
	s.router.Use(sentryMiddleware.Handle)

	s.router.Handle("/static/*", fs)
	s.router.Get("/{view:(?:view)}/{fileId}", s.downloadHandler)
	s.router.Get("/{fileId}", s.downloadHandler)

	s.protectedRouter = chi.NewRouter()
	s.protectedRouter.Use(middleware.Logger)
	s.protectedRouter.Use(middleware.Recoverer)
	s.protectedRouter.Use(limiter)

	if s.sitePasswordHash != "" {
		s.router.Get("/login", s.loginGETHandler)
		s.router.Post("/login", s.loginPOSTHandler)
		s.protectedRouter.Use(SitePasswordMiddleware(s.sitePasswordHash))
	}

	s.protectedRouter.Use(sentryMiddleware.Handle)

	s.protectedRouter.Get("/", s.indexHandler)
	s.protectedRouter.Post("/", s.formHandler)

	s.router.Mount("/", s.protectedRouter)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      s.router,
	}
	s.logger.Debug(fmt.Sprintf("Server started [storage=%s]", s.storage.Type()))

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
