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

func Info(format string, args ...interface{}) {
	fmt.Printf(blue+"INFO: "+reset+format+"\n", args...)
}

func Warn(format string, args ...interface{}) {
	fmt.Printf(yellow+"WARN: "+reset+format+"\n", args...)
}

func Error(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, red+"ERROR: "+reset+format+"\n", args...)
}

func Success(format string, args ...interface{}) {
	fmt.Printf(green+"SUCCESS: "+reset+format+"\n", args...)
}

func Fatal(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
