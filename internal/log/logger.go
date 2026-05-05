package log

import (
	stdlog "log"
	"os"
)

type Logger struct {
	logger *stdlog.Logger
}

func New() *Logger {
	return &Logger{logger: stdlog.New(os.Stdout, "gokv ", stdlog.LstdFlags|stdlog.Lmsgprefix)}
}

func (l *Logger) Printf(format string, args ...any) {
	l.logger.Printf(format, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.logger.Fatalf(format, args...)
}
