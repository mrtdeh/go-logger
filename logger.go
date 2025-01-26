package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

type Config struct {
	Plugin   string
	Category string
	Level    string
	Output   io.Writer
}

type Logger struct {
	zerolog.Logger
	loggers   map[string]*Logger
	loggersMu *sync.Mutex
	aftaLevel bool
}

func NewLogger() *Logger {
	return &Logger{
		loggers:   make(map[string]*Logger),
		loggersMu: &sync.Mutex{},
		aftaLevel: true,
	}
}

func (l *Logger) DisableAFTA() {
	l.aftaLevel = false
}

func (l *Logger) Init(config Config) *Logger {
	l.loggersMu.Lock()
	defer l.loggersMu.Unlock()

	if existingLogger, exists := l.loggers[config.Category]; exists {
		return existingLogger
	}

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var output zerolog.ConsoleWriter
	if config.Output == nil {
		filepath := fmt.Sprintf("/var/log/siem/%s/%s.log", config.Plugin, config.Category)

		runLogFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			panic(fmt.Errorf("failed to open log file: %w", err))
		}
		output = zerolog.ConsoleWriter{Out: runLogFile}

	} else {
		output = zerolog.ConsoleWriter{Out: config.Output}
	}

	newLogger := zerolog.New(output).With().Str("plugin", config.Plugin).Timestamp().Logger()
	levelObj := getLogLevel(config.Level)
	newLogger = newLogger.Level(levelObj)

	logger := &Logger{Logger: newLogger, loggers: l.loggers, loggersMu: l.loggersMu}
	l.loggers[config.Category] = logger

	return logger
}

func (l *Logger) GetLogger(category string) *Logger {
	l.loggersMu.Lock()
	defer l.loggersMu.Unlock()

	if logger, exists := l.loggers[category]; exists {
		return logger
	}
	return nil
}

func getLogLevel(level string) zerolog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
