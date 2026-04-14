package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Logger is the structured logger interface used across all layers.
// Backed by zerolog — zero-allocation JSON logging.
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	With(fields ...Field) Logger
	Flush() error
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// zeroLogger wraps zerolog.Logger behind the Logger interface
type zeroLogger struct {
	zl zerolog.Logger
}

// New creates a zerolog-backed Logger writing JSON to writer.
// zerolog is concurrent-safe — no mutex needed.
func New(level LogLevel, writer io.Writer) Logger {
	if writer == nil {
		writer = os.Stdout
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano

	zl := zerolog.New(writer).
		Level(toZeroLevel(level)).
		With().
		Timestamp().
		Logger()

	return &zeroLogger{zl: zl}
}

func (l *zeroLogger) Debug(msg string, fields ...Field) { l.emit(l.zl.Debug(), msg, fields...) }
func (l *zeroLogger) Info(msg string, fields ...Field)  { l.emit(l.zl.Info(), msg, fields...) }
func (l *zeroLogger) Warn(msg string, fields ...Field)  { l.emit(l.zl.Warn(), msg, fields...) }
func (l *zeroLogger) Error(msg string, fields ...Field) { l.emit(l.zl.Error(), msg, fields...) }

// With returns a new logger with the given fields attached to every subsequent entry
func (l *zeroLogger) With(fields ...Field) Logger {
	ctx := l.zl.With()
	for _, f := range fields {
		ctx = applyToContext(ctx, f)
	}
	return &zeroLogger{zl: ctx.Logger()}
}

// Flush is a no-op — zerolog writes synchronously
func (l *zeroLogger) Flush() error { return nil }

func (l *zeroLogger) emit(event *zerolog.Event, msg string, fields ...Field) {
	for _, f := range fields {
		event = applyToEvent(event, f)
	}
	event.Msg(msg)
}

// applyToEvent adds a Field to a zerolog event using the correct typed method.
// Using typed methods (Str, Int, etc.) avoids reflection and keeps allocations near zero.
func applyToEvent(e *zerolog.Event, f Field) *zerolog.Event {
	switch v := f.Value.(type) {
	case string:
		return e.Str(f.Key, v)
	case int:
		return e.Int(f.Key, v)
	case int64:
		return e.Int64(f.Key, v)
	case float64:
		return e.Float64(f.Key, v)
	case bool:
		return e.Bool(f.Key, v)
	default:
		return e.Interface(f.Key, v)
	}
}

// applyToContext adds a Field to a zerolog context (used by With())
func applyToContext(ctx zerolog.Context, f Field) zerolog.Context {
	switch v := f.Value.(type) {
	case string:
		return ctx.Str(f.Key, v)
	case int:
		return ctx.Int(f.Key, v)
	case int64:
		return ctx.Int64(f.Key, v)
	case float64:
		return ctx.Float64(f.Key, v)
	case bool:
		return ctx.Bool(f.Key, v)
	default:
		return ctx.Interface(f.Key, v)
	}
}

func toZeroLevel(l LogLevel) zerolog.Level {
	switch l {
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelInfo:
		return zerolog.InfoLevel
	case LevelWarn:
		return zerolog.WarnLevel
	case LevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Field constructor helpers — identical signatures to before, callers unchanged

func String(key, value string) Field          { return Field{Key: key, Value: value} }
func Int(key string, value int) Field         { return Field{Key: key, Value: value} }
func Int64(key string, value int64) Field     { return Field{Key: key, Value: value} }
func Duration(key string, d time.Duration) Field { return Field{Key: key, Value: d.String()} }
func Any(key string, value interface{}) Field { return Field{Key: key, Value: value} }

// Err creates an error field. Nil-safe.
func Err(err error) Field {
	if err == nil {
		return Field{Key: "error", Value: nil}
	}
	return Field{Key: "error", Value: err.Error()}
}
