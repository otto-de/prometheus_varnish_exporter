package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

// SlogErrorLogger adapts slog.Logger to the interface expected by the prometheus client library for logging errors.
type SlogErrorLogger struct {
	logger *slog.Logger
}

func (l *SlogErrorLogger) Println(v ...interface{}) {
	l.logger.Error(fmt.Sprint(v...))
}

// initSlogger overwrites the global default slog instance
func initSlogger(logLevel string, json bool) {
	var handler slog.Handler
	var level slog.Level
	switch strings.ToUpper(logLevel) {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	if json {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}
	slog.SetDefault(slog.New(handler))
}

func logFatal(message string, args ...interface{}) {
	slog.Error(message, args...)
	os.Exit(1)
}

func logFatalError(err error) {
	if err != nil {
		logFatal(err.Error())
	}
}

// strings

type caseSensitivity int

const (
	caseSensitive   caseSensitivity = 0
	caseInsensitive caseSensitivity = 1
)

func startsWith(str, prefix string, cs caseSensitivity) bool {
	if cs == caseSensitive {
		return strings.HasPrefix(str, prefix)
	}
	return strings.HasPrefix(strings.ToLower(str), strings.ToLower(prefix))
}

func startsWithAny(str string, prefixes []string, cs caseSensitivity) bool {
	for _, prefix := range prefixes {
		if startsWith(str, prefix, cs) {
			return true
		}
	}
	return false
}

func endsWith(str, postfix string, cs caseSensitivity) bool {
	if cs == caseSensitive {
		return strings.HasSuffix(str, postfix)
	}
	return strings.HasSuffix(strings.ToLower(str), strings.ToLower(postfix))
}

func endsWithAny(str string, postfixes []string, cs caseSensitivity) bool {
	for _, postfix := range postfixes {
		if endsWith(str, postfix, cs) {
			return true
		}
	}
	return false
}

// file

// Returns if file/dir in path exists.
func fileExists(path string) bool {
	if len(path) == 0 {
		return false
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// data

func stringProperty(data map[string]interface{}, key string) (string, error) {
	if value, ok := data[key]; ok {
		if vStr, ok := value.(string); ok {
			return vStr, nil
		} else {
			return "", fmt.Errorf("%s is not a string", key)
		}
	}
	return "", nil
}
