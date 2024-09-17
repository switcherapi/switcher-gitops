package utils

import (
	"fmt"
	"log"

	"github.com/switcherapi/switcher-gitops/src/config"
)

const (
	LogLevelInfo  = "INFO"
	LogLevelError = "ERROR"
	LogLevelDebug = "DEBUG"
)

func LogInfo(message string, args ...interface{}) {
	Log(LogLevelInfo, message, args...)
}

func LogError(message string, args ...interface{}) {
	Log(LogLevelError, message, args...)
}

func LogDebug(message string, args ...interface{}) {
	Log(LogLevelDebug, message, args...)
}

func Log(logLevel string, message string, args ...interface{}) {
	currentLogLevel := config.GetEnv("LOG_LEVEL")

	if currentLogLevel == LogLevelDebug || currentLogLevel == LogLevelError ||
		currentLogLevel == logLevel || LogLevelError == logLevel {
		log.Printf("[%s] %s\n", logLevel, fmt.Sprintf(message, args...))
	}
}
