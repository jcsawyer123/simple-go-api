package logger

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

var logger zerolog.Logger

func Init() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		FormatMessage: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return fmt.Sprintf("%s", i)
		},
		FormatFieldName: func(i interface{}) string {
			return ""
		},
		FormatFieldValue: func(i interface{}) string {
			if i == "null" {
				return ""
			}
			return fmt.Sprintf("%s", i)
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return fmt.Sprintf("%s", i)
		},
	}
	logger = zerolog.New(output).With().Timestamp().Caller().Logger()
}

func getRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if reqID, ok := ctx.Value(middleware.RequestIDKey).(string); ok {
		return fmt.Sprintf("[%s] ", reqID)
	}
	return ""
}

// Structured logging methods
func Debug() *zerolog.Event {
	return logger.Debug()
}

func DebugWCtx(ctx context.Context) *zerolog.Event {
	return logger.Debug().Str("prefix", getRequestID(ctx))
}

func Info() *zerolog.Event {
	return logger.Info()
}

func InfoWCtx(ctx context.Context) *zerolog.Event {
	return logger.Info().Str("prefix", getRequestID(ctx))
}

func Warn() *zerolog.Event {
	return logger.Warn()
}

func WarnWCtx(ctx context.Context) *zerolog.Event {
	return logger.Warn().Str("prefix", getRequestID(ctx))
}

func Error() *zerolog.Event {
	return logger.Error()
}

func ErrorWCtx(ctx context.Context) *zerolog.Event {
	return logger.Error().Str("prefix", getRequestID(ctx))
}

func Fatal() *zerolog.Event {
	return logger.Fatal()
}

func FatalWCtx(ctx context.Context) *zerolog.Event {
	return logger.Fatal().Str("prefix", getRequestID(ctx))
}

// Printf style convenience methods
func Debugf(format string, v ...interface{}) {
	logger.Debug().Msg(fmt.Sprintf(format, v...))
}

func DebugfWCtx(ctx context.Context, format string, v ...interface{}) {
	logger.Debug().Msg(getRequestID(ctx) + fmt.Sprintf(format, v...))
}

func Infof(format string, v ...interface{}) {
	logger.Info().Msg(fmt.Sprintf(format, v...))
}

func InfofWCtx(ctx context.Context, format string, v ...interface{}) {
	logger.Info().Msg(getRequestID(ctx) + fmt.Sprintf(format, v...))
}

func Warnf(format string, v ...interface{}) {
	logger.Warn().Msg(fmt.Sprintf(format, v...))
}

func WarnCTXf(ctx context.Context, format string, v ...interface{}) {
	logger.Warn().Msg(getRequestID(ctx) + fmt.Sprintf(format, v...))
}

func Errorf(format string, v ...interface{}) {
	logger.Error().Msg(fmt.Sprintf(format, v...))
}

func ErrorfWCtx(ctx context.Context, format string, v ...interface{}) {
	logger.Error().Msg(getRequestID(ctx) + fmt.Sprintf(format, v...))
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatal().Msg(fmt.Sprintf(format, v...))
}

func FatalfWCtx(ctx context.Context, format string, v ...interface{}) {
	logger.Fatal().Msg(getRequestID(ctx) + fmt.Sprintf(format, v...))
}
