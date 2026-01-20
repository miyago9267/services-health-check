package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

type Config struct {
	Level  string
	Format string
	Output io.Writer
}

type Logger struct {
	mu    sync.Mutex
	level Level
	out   io.Writer
	json  bool
}

func New(cfg Config) *Logger {
	lvl := parseLevel(cfg.Level)
	out := cfg.Output
	if out == nil {
		out = os.Stdout
	}
	return &Logger{
		level: lvl,
		out:   out,
		json:  strings.EqualFold(strings.TrimSpace(cfg.Format), "json"),
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	l.logf(LevelDebug, "DEBUG", format, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.logf(LevelInfo, "INFO", format, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.logf(LevelWarn, "WARN", format, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.logf(LevelError, "ERROR", format, args...)
}

func (l *Logger) logf(level Level, label string, format string, args ...any) {
	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, args...)
	timestamp := time.Now().Format(time.RFC3339)

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.json {
		payload := map[string]string{
			"time":  timestamp,
			"level": label,
			"msg":   msg,
		}
		enc, err := json.Marshal(payload)
		if err == nil {
			fmt.Fprintf(l.out, "%s\n", enc)
			return
		}
	}
	fmt.Fprintf(l.out, "%s [%s] %s\n", timestamp, label, msg)
}

func parseLevel(raw string) Level {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return LevelDebug
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	default:
		return LevelInfo
	}
}
