package logger

import "log"

type ILogger interface {
	Debug(template string, args ...any)
	Info(template string, args ...any)
	Warn(template string, args ...any)
	Error(template string, args ...any)
	Fatal(template string, args ...any)
}

type defaultLogger struct {
}

func DefaultLogger() ILogger {
	return &defaultLogger{}
}

func (l *defaultLogger) Debug(template string, args ...any) {
	log.Printf(template, args...)
}

func (l *defaultLogger) Info(template string, args ...any) {
	log.Printf(template, args...)
}

func (l *defaultLogger) Warn(template string, args ...any) {
	log.Printf(template, args...)
}

func (l *defaultLogger) Error(template string, args ...any) {
	log.Printf(template, args...)
}

func (l *defaultLogger) Fatal(template string, args ...any) {
	log.Printf(template, args...)
}
