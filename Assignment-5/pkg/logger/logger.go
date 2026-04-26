package logger

import (
	"fmt"
	"log"
)

type Interface interface {
	Info(args ...interface{})
	Error(args ...interface{})
}

type Logger struct{}

func New() *Logger {
	return &Logger{}
}

func (l *Logger) Info(args ...interface{}) {
	log.Println("[INFO]", fmt.Sprint(args...))
}

func (l *Logger) Error(args ...interface{}) {
	log.Println("[ERROR]", fmt.Sprint(args...))
}
