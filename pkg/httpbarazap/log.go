package httpbarazap

import (
	"github.com/gopybara/httpbara"
	"go.uber.org/zap"
)

type zapLogger struct {
	log *zap.Logger
}

func New(log *zap.Logger) httpbara.Logger {
	return &zapLogger{log: log}
}

func (l *zapLogger) Info(message string, args ...any) {
	l.log.Info(message, l.mapFields(args...)...)
}

func (l *zapLogger) Debug(message string, args ...any) {
	l.log.Debug(message, l.mapFields(args...)...)
}

func (l *zapLogger) Error(message string, args ...any) {
	l.log.Error(message, l.mapFields(args...)...)
}

func (l *zapLogger) Panic(message string, args ...any) {
	l.log.Panic(message, l.mapFields(args...)...)
}

func (l *zapLogger) Warn(message string, args ...any) {
	l.log.Warn(message, l.mapFields(args...)...)
}

func (l *zapLogger) mapFields(fields ...any) []zap.Field {
	expectingKey := true
	result := make([]zap.Field, 0)
	key := ""

	for i := 0; i < len(fields); i++ {
		switch field := fields[i].(type) {
		case zap.Field:
			result = append(result, field)
		default:
			if expectingKey {
				key = field.(string)
			} else {
				switch field.(type) {
				case string:
					field = zap.String(key, field.(string))
				case int:
					field = zap.Int(key, field.(int))
				case int64:
					field = zap.Int64(key, field.(int64))
				case float64:
					field = zap.Float64(key, field.(float64))
				default:
					field = zap.Any(key, field)
				}

				key = ""
			}

			expectingKey = !expectingKey
		}
	}

	return result
}
