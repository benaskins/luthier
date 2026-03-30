// Package logger provides a configurable, leveled logger for maestro.
// It writes plain-text messages to an io.Writer, making it suitable for
// CLI tools that need human-readable stderr output without timestamps.
package logger

import (
	"fmt"
	"io"
	"os"
)

// Level controls which messages are emitted. Lower values are more verbose.
type Level int

const (
	// LevelDebug emits all messages including fine-grained execution details.
	LevelDebug Level = iota
	// LevelInfo emits informational progress messages and above. Default.
	LevelInfo
	// LevelWarn emits warnings and errors only.
	LevelWarn
	// LevelError emits errors only.
	LevelError
	// LevelSilent suppresses all output.
	LevelSilent
)

// Logger writes leveled messages to an io.Writer.
// The zero value is not usable; use New or Default.
type Logger struct {
	level Level
	w     io.Writer
}

// New returns a Logger at the given level writing to w.
// If w is nil, os.Stderr is used.
func New(level Level, w io.Writer) *Logger {
	if w == nil {
		w = os.Stderr
	}
	return &Logger{level: level, w: w}
}

// Default returns an Info-level logger that writes to os.Stderr.
func Default() *Logger {
	return New(LevelInfo, os.Stderr)
}

// Debug writes a message at debug level.
// The format and args follow fmt.Fprintf conventions.
func (l *Logger) Debug(format string, args ...any) {
	if l.level <= LevelDebug {
		fmt.Fprintf(l.w, format, args...)
	}
}

// Info writes a message at info level.
func (l *Logger) Info(format string, args ...any) {
	if l.level <= LevelInfo {
		fmt.Fprintf(l.w, format, args...)
	}
}

// Warn writes a message at warn level, prefixed with "warn: ".
func (l *Logger) Warn(format string, args ...any) {
	if l.level <= LevelWarn {
		fmt.Fprintf(l.w, "warn: "+format, args...)
	}
}

// Error writes a message at error level, prefixed with "error: ".
func (l *Logger) Error(format string, args ...any) {
	if l.level <= LevelError {
		fmt.Fprintf(l.w, "error: "+format, args...)
	}
}

// IsDebug reports whether debug-level messages are enabled.
// Useful to guard expensive formatting operations.
func (l *Logger) IsDebug() bool { return l.level <= LevelDebug }

// Level returns the configured log level.
func (l *Logger) Level() Level { return l.level }
