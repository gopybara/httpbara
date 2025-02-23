package httpbara

import (
	"fmt"
	"strings"
	"time"
)

type Logger interface {
	Info(message string, args ...any)
	Debug(message string, args ...any)
	Error(message string, args ...any)
	Panic(message string, args ...any)
	Warn(message string, args ...any)
}

type fmtLogger struct {
	Logger
}

func (l *fmtLogger) mapFields(fields ...any) string {
	if len(fields)%2 != 0 {
		return ""
	}

	var sb strings.Builder

	for i := 0; i < len(fields); i += 2 {
		if i > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString(fmt.Sprintf("%v=%v", fields[i], fields[i+1]))
	}

	return sb.String()
}

func (l *fmtLogger) log(level string, message string, args ...any) {
	timestamp := time.Now().Format(time.RFC3339)
	fields := l.mapFields(args...)

	if fields != "" {
		fmt.Printf("%s [%s] %s %s\n", timestamp, level, message, fields)
	} else {
		fmt.Printf("%s [%s] %s\n", timestamp, level, message)
	}
}

func (l *fmtLogger) Info(message string, args ...any) {
	l.log("INFO", message, args...)
}

func (l *fmtLogger) Debug(message string, args ...any) {
	l.log("DEBUG", message, args...)
}

func (l *fmtLogger) Error(message string, args ...any) {
	l.log("ERROR", message, args...)
}

func (l *fmtLogger) Panic(message string, args ...any) {
	l.log("PANIC", message, args...)
	panic(message)
}

func (l *fmtLogger) Warn(message string, args ...any) {
	l.log("WARN", message, args...)
}

func NewFmtLogger() Logger {
	return &fmtLogger{}
}
