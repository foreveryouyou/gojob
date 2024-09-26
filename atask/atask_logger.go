package atask

import "github.com/foreveryouyou/gojob/pkg/logger"

// asynqLogger 自定义asynqLogger
type asynqLogger struct {
	logger logger.ILogger
}

func (l *asynqLogger) formatArgs(args ...any) (msg string, _args []any) {
	if len(args) > 0 {
		msg = args[0].(string)
		_args = args[1:]
		return
	}

	return "", args
}

func (l *asynqLogger) Debug(args ...any) {
	msg, _args := l.formatArgs(args...)
	l.logger.Debug(msg, _args...)
}

func (l *asynqLogger) Info(args ...any) {
	msg, _args := l.formatArgs(args...)
	l.logger.Info(msg, _args...)
}

func (l *asynqLogger) Warn(args ...any) {
	msg, _args := l.formatArgs(args...)
	l.logger.Warn(msg, _args...)
}

func (l *asynqLogger) Error(args ...any) {
	msg, _args := l.formatArgs(args...)
	l.logger.Error(msg, _args...)
}

func (l *asynqLogger) Fatal(args ...any) {
	msg, _args := l.formatArgs(args...)
	l.logger.Fatal(msg, _args...)
}
