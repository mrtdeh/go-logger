package logger

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
)

type Logger struct {
	zerolog.Logger
}

var loggers = make(map[string]*Logger)

func Init(plugin, category, level string) *Logger {

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Default level for this example is info, unless debug flag is present

	runLogFile, err := os.OpenFile(
		fmt.Sprintf("/var/log/siem/%s/%s.log", plugin, category),
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		panic(err)
	}
	// initialize logger
	l := zerolog.New(runLogFile).With().Str("plugin", plugin).Timestamp().Logger()
	levelObj := getLevel(level)
	l = l.Level(levelObj)

	ls := &Logger{l}
	loggers[category] = ls

	return ls
}

func GetLogger(category string) *Logger {
	if l, ok := loggers[category]; ok {
		return l
	}
	return nil
}

func getLevel(l string) zerolog.Level {

	switch l {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "error":
		return zerolog.ErrorLevel
	}
	return zerolog.Disabled
}
