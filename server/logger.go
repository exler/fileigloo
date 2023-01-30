package server

import (
	"log"
	"os"

	"github.com/getsentry/sentry-go"
	colors "github.com/logrusorgru/aurora"
)

type Logger struct {
	*log.Logger

	sentryEnabled bool
}

func NewLogger(sentryEnabled bool) *Logger {
	return &Logger{
		Logger:        log.New(os.Stderr, colors.Blue("[fileigloo] ").String(), log.LstdFlags),
		sentryEnabled: sentryEnabled,
	}
}

func (l *Logger) Error(err error) {
	l.Logger.Print(colors.Red(err.Error()).String())
	if l.sentryEnabled {
		sentry.CaptureException(err)
	}
}

func (l *Logger) Info(msg string) {
	l.Logger.Print(msg)
	if l.sentryEnabled {
		sentry.CaptureMessage(msg)
	}
}
