package logger

import (
	"log"
	"os"

	colors "github.com/logrusorgru/aurora/v4"
)

type Logger struct {
	*log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		Logger: log.New(os.Stderr, "[fileigloo] ", log.LstdFlags),
	}
}

func (l *Logger) Error(err error) {
	l.Logger.Print(colors.Red(err.Error()).String())
}

func (l *Logger) Info(msg string) {
	l.Logger.Print(msg)
}

func (l *Logger) Debug(msg string) {
	l.Logger.Print(colors.Gray(10, msg).String())
}
