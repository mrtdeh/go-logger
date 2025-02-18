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

type MyLogger struct {
	zerolog.Logger
	loggersMu *sync.Mutex
	aftaLevel bool
}

func NewLogger() *MyLogger {
	return &MyLogger{
		loggersMu: &sync.Mutex{},
		aftaLevel: true,
	}
}

func (l *MyLogger) AFTA() *AftaLogger {
	af := &AftaLogger{logger: l}
	af.logger.Logger = af.logger.Logger.With().Str("type", "afta").Logger()
	return af
}

func (l *MyLogger) DisableAFTA() {
	l.aftaLevel = false
}

func (l *MyLogger) Init(config Config) *MyLogger {
	l.loggersMu.Lock()
	defer l.loggersMu.Unlock()

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var output io.Writer
	if config.Output == nil {
		filepath := fmt.Sprintf("/var/log/siem/%s/%s.log", config.Plugin, config.Category)

		runLogFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			panic(fmt.Errorf("failed to open log file: %w", err))
		}
		output = runLogFile

	} else {
		output = config.Output
	}

	newLogger := zerolog.New(output).With().Str("plugin", config.Plugin).Timestamp().Logger()
	levelObj := getLogLevel(config.Level)
	newLogger = newLogger.Level(levelObj)

	l.Logger = newLogger
	return l
}

// func (l *MyLogger) GetLogger(category string) *MyLogger {
// 	l.loggersMu.Lock()
// 	defer l.loggersMu.Unlock()

// 	if logger, exists := l.loggers[category]; exists {
// 		return logger
// 	}
// 	return nil
// }

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
