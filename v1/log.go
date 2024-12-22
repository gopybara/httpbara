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

	mapFields(fields ...any) string
}

type FmtLogger struct {
	Logger
}

func (l *FmtLogger) mapFields(fields ...any) string {
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

func (l *FmtLogger) log(level string, message string, args ...any) {
	timestamp := time.Now().Format(time.RFC3339)
	fields := l.mapFields(args...)
	if fields != "" {
		fmt.Printf("%s [%s] %s %s\n", timestamp, level, message, fields)
	} else {
		fmt.Printf("%s [%s] %s\n", timestamp, level, message)
	}
}

func (l *FmtLogger) Info(message string, args ...any) {
	l.log("INFO", message, args...)
}

func (l *FmtLogger) Debug(message string, args ...any) {
	l.log("DEBUG", message, args...)
}

func (l *FmtLogger) Error(message string, args ...any) {
	l.log("ERROR", message, args...)
}

func (l *FmtLogger) Panic(message string, args ...any) {
	l.log("PANIC", message, args...)
	panic(message)
}

func (l *FmtLogger) Warn(message string, args ...any) {
	l.log("WARN", message, args...)
}

func NewFmtLogger() *FmtLogger {
	return &FmtLogger{}
}
