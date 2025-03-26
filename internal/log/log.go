package log

import (
	"fmt"
	"os"
	"sync"

	"github.com/statico/llmscript/internal/progress"
	"golang.org/x/term"
)

const (
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	reset  = "\033[0m"
)

var (
	level     = InfoLevel
	levelLock sync.RWMutex
	spinner   *progress.Spinner
	spinnerMu sync.Mutex
)

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

func SetLevel(l Level) {
	levelLock.Lock()
	defer levelLock.Unlock()
	level = l
}

func getLevel() Level {
	levelLock.RLock()
	defer levelLock.RUnlock()
	return level
}

func isTTY() bool {
	return term.IsTerminal(int(os.Stderr.Fd()))
}

func updateSpinner(format string, args ...interface{}) {
	if !isTTY() {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
		return
	}

	spinnerMu.Lock()
	defer spinnerMu.Unlock()

	if spinner == nil {
		spinner = progress.NewSpinner(fmt.Sprintf(format, args...))
	} else {
		spinner.SetMessage(fmt.Sprintf(format, args...))
	}
	fmt.Fprintf(os.Stderr, "\r%s", spinner)
}

func Spinner(format string, args ...interface{}) {
	updateSpinner(format, args...)
}

func Info(format string, args ...interface{}) {
	if getLevel() <= InfoLevel {
		if getLevel() == DebugLevel {
			fmt.Fprintf(os.Stderr, blue+"INFO: "+reset+format+"\n", args...)
		} else if isTTY() {
			updateSpinner(format, args...)
		}
	}
}

func Debug(format string, args ...interface{}) {
	if getLevel() <= DebugLevel {
		fmt.Fprintf(os.Stderr, blue+"DEBUG: "+reset+format+"\n", args...)
	}
}

func Warn(format string, args ...interface{}) {
	if getLevel() <= WarnLevel {
		fmt.Fprintf(os.Stderr, yellow+"WARN: "+reset+format+"\n", args...)
	}
}

func Error(format string, args ...interface{}) {
	if getLevel() <= ErrorLevel {
		fmt.Fprintf(os.Stderr, red+"ERROR: "+reset+format+"\n", args...)
	}
}

func Fatal(format string, args ...interface{}) {
	if spinner != nil {
		spinner.Stop()
		fmt.Fprintf(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr, red+"FATAL: "+reset+format+"\n", args...)
	os.Exit(1)
}

func Success(format string, args ...interface{}) {
	if spinner != nil {
		spinner.Stop()
		fmt.Fprintf(os.Stderr, "\n")
	}
	fmt.Fprintf(os.Stderr, green+"✓ "+reset+format+"\n", args...)
}
