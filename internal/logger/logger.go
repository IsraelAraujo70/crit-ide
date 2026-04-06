package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger writes timestamped messages to a log file.
// When disabled (default), all writes are silently discarded.
type Logger struct {
	mu      sync.Mutex
	w       io.Writer
	file    *os.File
	enabled bool
}

// global is the package-level logger instance.
var global = &Logger{w: io.Discard}

// Init enables debug logging to a file in the given directory.
// Call this once at startup when --debug is passed.
// The log file is created at <dir>/crit-ide-<date>.log.
func Init(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("logger: create dir: %w", err)
	}

	name := fmt.Sprintf("crit-ide-%s.log", time.Now().Format("2006-01-02"))
	path := filepath.Join(dir, name)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("logger: open file: %w", err)
	}

	global.mu.Lock()
	global.file = f
	global.w = f
	global.enabled = true
	global.mu.Unlock()

	Info("logging started: %s", path)
	return nil
}

// Close flushes and closes the log file.
func Close() {
	global.mu.Lock()
	defer global.mu.Unlock()
	if global.file != nil {
		_ = global.file.Close()
		global.file = nil
	}
	global.w = io.Discard
	global.enabled = false
}

// Enabled returns true if debug logging is active.
func Enabled() bool {
	return global.enabled
}

// Writer returns the underlying io.Writer (the log file or io.Discard).
// Useful for redirecting subprocess stderr.
func Writer() io.Writer {
	return global.w
}

// Info logs an informational message.
func Info(format string, args ...any) {
	write("INFO", format, args...)
}

// Warn logs a warning message.
func Warn(format string, args ...any) {
	write("WARN", format, args...)
}

// Error logs an error message.
func Error(format string, args ...any) {
	write("ERR ", format, args...)
}

// Debug logs a debug message (only visible when logging is enabled).
func Debug(format string, args ...any) {
	write("DEBG", format, args...)
}

func write(level, format string, args ...any) {
	global.mu.Lock()
	defer global.mu.Unlock()
	if !global.enabled {
		return
	}
	ts := time.Now().Format("15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(global.w, "%s [%s] %s\n", ts, level, msg)
}
