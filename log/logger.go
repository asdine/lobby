package log

import (
	"io"
	"log"
)

type Logger struct {
	prefix       string
	logger       *log.Logger
	DebugEnabled bool
}

func New(out io.Writer, prefix string) *Logger {
	return &Logger{
		logger: log.New(out, "", log.Flags()),
		prefix: prefix,
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
	if !l.DebugEnabled {
		return
	}

	l.leveledPrintln("d |", v...)
}

func (l *Logger) Debugf(format string, v ...interface{}) {
	if !l.DebugEnabled {
		return
	}

	l.leveledPrintf("d | ", format, v...)
}
