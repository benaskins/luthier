package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/benaskins/maestro/internal/logger"
)

func TestNew_WritesToConfiguredWriter(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelInfo, &buf)
	l.Info("hello %s\n", "world")
	if got := buf.String(); got != "hello world\n" {
		t.Errorf("Info output = %q, want %q", got, "hello world\n")
	}
}

func TestNew_NilWriterUsesStderr(t *testing.T) {
	// Just verify it doesn't panic.
	l := logger.New(logger.LevelInfo, nil)
	l.Info("should not panic\n")
}

func TestDefault_IsInfoLevel(t *testing.T) {
	l := logger.Default()
	if l.Level() != logger.LevelInfo {
		t.Errorf("Default level = %v, want LevelInfo", l.Level())
	}
}

func TestDebug_SuppressedAtInfoLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelInfo, &buf)
	l.Debug("should not appear\n")
	if buf.Len() != 0 {
		t.Errorf("expected no output at LevelInfo for Debug, got %q", buf.String())
	}
}

func TestDebug_EmittedAtDebugLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelDebug, &buf)
	l.Debug("debug message\n")
	if !strings.Contains(buf.String(), "debug message") {
		t.Errorf("expected debug message, got %q", buf.String())
	}
}

func TestWarn_SuppressedAtErrorLevel(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelError, &buf)
	l.Warn("should not appear\n")
	if buf.Len() != 0 {
		t.Errorf("expected no warn output at LevelError, got %q", buf.String())
	}
}

func TestWarn_HasPrefix(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelWarn, &buf)
	l.Warn("something wrong\n")
	if !strings.HasPrefix(buf.String(), "warn: ") {
		t.Errorf("Warn output = %q, want it to start with 'warn: '", buf.String())
	}
}

func TestError_HasPrefix(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelError, &buf)
	l.Error("fatal condition\n")
	if !strings.HasPrefix(buf.String(), "error: ") {
		t.Errorf("Error output = %q, want it to start with 'error: '", buf.String())
	}
}

func TestSilent_SuppressesAll(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelSilent, &buf)
	l.Debug("debug\n")
	l.Info("info\n")
	l.Warn("warn\n")
	l.Error("error\n")
	if buf.Len() != 0 {
		t.Errorf("LevelSilent should suppress all output, got %q", buf.String())
	}
}

func TestIsDebug_TrueAtDebugLevel(t *testing.T) {
	l := logger.New(logger.LevelDebug, nil)
	if !l.IsDebug() {
		t.Error("IsDebug() = false at LevelDebug, want true")
	}
}

func TestIsDebug_FalseAtInfoLevel(t *testing.T) {
	l := logger.New(logger.LevelInfo, nil)
	if l.IsDebug() {
		t.Error("IsDebug() = true at LevelInfo, want false")
	}
}

func TestLevelOrdering_InfoSuppressesDebug(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelInfo, &buf)

	l.Debug("no\n")
	l.Info("yes\n")
	l.Warn("yes\n")
	l.Error("yes\n")

	out := buf.String()
	if strings.Contains(out, "no") {
		t.Errorf("Debug message should be suppressed at LevelInfo, got %q", out)
	}
	if !strings.Contains(out, "yes") {
		t.Errorf("Info/Warn/Error messages should appear at LevelInfo, got %q", out)
	}
}

func TestLevelOrdering_WarnSuppressesDebugAndInfo(t *testing.T) {
	var buf bytes.Buffer
	l := logger.New(logger.LevelWarn, &buf)

	l.Debug("debug\n")
	l.Info("info\n")
	l.Warn("warn\n")
	l.Error("error\n")

	out := buf.String()
	if strings.Contains(out, "debug") {
		t.Errorf("Debug suppressed at LevelWarn, got %q", out)
	}
	if strings.Contains(out, "info") {
		t.Errorf("Info suppressed at LevelWarn, got %q", out)
	}
	if !strings.Contains(out, "warn") {
		t.Errorf("Warn visible at LevelWarn, got %q", out)
	}
	if !strings.Contains(out, "error") {
		t.Errorf("Error visible at LevelWarn, got %q", out)
	}
}
