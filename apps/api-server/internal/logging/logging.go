package logging

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Logger

// Config holds logging configuration
type Config struct {
	Level  string
	Format string // "json" or "text"
	Output string // "stdout" or "stderr"
}

// DefaultConfig returns default logging configuration
func DefaultConfig() Config {
	return Config{
		Level:  os.Getenv("LOG_LEVEL"),
		Format: os.Getenv("LOG_FORMAT"),
		Output: os.Getenv("LOG_OUTPUT"),
	}
}

// Init initializes the global logger with the given configuration
func Init(config Config) error {
	Logger = logrus.New()

	// Set log level
	if config.Level == "" {
		config.Level = "info"
	}
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	Logger.SetLevel(level)

	// Set log format
	if config.Format == "" {
		config.Format = "json"
	}
	if config.Format == "json" {
		Logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "function",
			},
		})
	} else {
		Logger.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: time.RFC3339,
			FullTimestamp:   true,
		})
	}

	// Set output
	if config.Output == "" || config.Output == "stdout" {
		Logger.SetOutput(os.Stdout)
	} else {
		Logger.SetOutput(os.Stderr)
	}

	return nil
}

// InitWithDefaults initializes the global logger with default configuration
func InitWithDefaults() error {
	return Init(DefaultConfig())
}

// WithField returns a logger with a single field
func WithField(key string, value any) *logrus.Entry {
	return Logger.WithField(key, value)
}

// WithFields returns a logger with multiple fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return Logger.WithFields(fields)
}

// WithError returns a logger with an error field
func WithError(err error) *logrus.Entry {
	return Logger.WithError(err)
}

// Info logs an info message
func Info(args ...any) {
	Logger.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...any) {
	Logger.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...any) {
	Logger.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...any) {
	Logger.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...any) {
	Logger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...any) {
	Logger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...any) {
	Logger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...any) {
	Logger.Fatalf(format, args...)
}

// Debug logs a debug message
func Debug(args ...any) {
	Logger.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...any) {
	Logger.Debugf(format, args...)
}
