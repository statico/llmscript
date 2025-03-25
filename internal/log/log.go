package log

import (
	"fmt"
	"os"
)

const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var currentLevel = InfoLevel

func SetLevel(level LogLevel) {
	currentLevel = level
}

func shouldLog(level LogLevel) bool {
	return level >= currentLevel
}

func Debug(format string, args ...interface{}) {
	if shouldLog(DebugLevel) {
		fmt.Fprintf(os.Stderr, blue+"DEBUG: "+reset+format+"\n", args...)
	}
}

func Info(format string, args ...interface{}) {
	if shouldLog(InfoLevel) {
		fmt.Fprintf(os.Stderr, blue+"INFO: "+reset+format+"\n", args...)
	}
}

func Warn(format string, args ...interface{}) {
	if shouldLog(WarnLevel) {
		fmt.Fprintf(os.Stderr, yellow+"WARN: "+reset+format+"\n", args...)
	}
}

func Error(format string, args ...interface{}) {
	if shouldLog(ErrorLevel) {
		fmt.Fprintf(os.Stderr, red+"ERROR: "+reset+format+"\n", args...)
	}
}

func Success(format string, args ...interface{}) {
	if shouldLog(InfoLevel) {
		fmt.Fprintf(os.Stderr, green+"SUCCESS: "+reset+format+"\n", args...)
	}
}

func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
