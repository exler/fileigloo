package logger

import (
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	colors "github.com/logrusorgru/aurora/v4"
)

type Logger struct {
	*log.Logger

	sentryHub *sentry.Hub
}

func NewLogger(sentry_dsn string) *Logger {
	var sentryHub *sentry.Hub
	if sentry_dsn != "" {
		var sentryClient *sentry.Client
		var err error
		if sentryClient, err = sentry.NewClient(sentry.ClientOptions{Dsn: sentry_dsn}); err != nil {
			log.Fatal(err)
		}
		sentryHub = sentry.NewHub(sentryClient, sentry.NewScope())
	}

	return &Logger{
		Logger:    log.New(os.Stderr, colors.Blue("[fileigloo] ").String(), log.LstdFlags),
		sentryHub: sentryHub,
	}
}

func (l *Logger) Error(err error) {
	l.Logger.Print(colors.Red(err.Error()).String())
	if l.sentryHub != nil {
		l.sentryHub.CaptureException(err)
	}
}

func (l *Logger) Info(msg string) {
	l.Logger.Print(msg)
	if l.sentryHub != nil {
		l.sentryHub.CaptureMessage(msg)
	}
}