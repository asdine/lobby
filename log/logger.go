package log

import (
	"io"
	"log"
	"os"
)

type Logger struct {
	prefix       string
	logger       *log.Logger
	debugEnabled bool
}

func New(opts ...func(*Logger)) *Logger {
	var l Logger

	for _, o := range opts {
		o(&l)
	}

	if l.logger == nil {
		Output(os.Stderr)(&l)
	}

	return &l
}

func Prefix(prefix string) func(*Logger) {
	return func(l *Logger) {
		l.prefix = prefix
	}
}

func Debug(debug bool) func(*Logger) {
	return func(l *Logger) {
		l.debugEnabled = debug
	}
}

func Output(out io.Writer) func(*Logger) {
	return func(l *Logger) {
		l.logger = log.New(out, "", log.Flags())
	}
}

func StdLogger(lg *log.Logger) func(*Logger) {
	return func(l *Logger) {
		l.logger = lg
	}
}

func (l *Logger) Println(v ...interface{}) {
	l.leveledPrintln("i |", v...)
}

func (l *Logger) leveledPrintln(level string, v ...interface{}) {
	if l.prefix != "" {
		v = append([]interface{}{level, l.prefix}, v...)
	} else {
		v = append([]interface{}{level}, v...)
	}

	l.logger.Println(v...)
}

func (l *Logger) leveledPrintf(level string, format string, v ...interface{}) {
	if l.prefix != "" {
		format = level + l.prefix + " " + format
	} else {
		format = level + format
	}

	l.logger.Printf(format, v...)
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.leveledPrintf("i | ", format, v...)
}

func (l *Logger) Debug(v ...interface{}) {
	if !l.debugEnabled {
		return
	}

	l.leveledPrintln("d |", v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if !l.debugEnabled {
		return
	}

	l.leveledPrintf("d | ", format, v...)
}
