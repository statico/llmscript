package progress

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

type Spinner struct {
	message      atomic.Value
	messageWidth int
	parts        []string
	value        int
	ticker       *time.Ticker
	started      time.Time
	stopped      time.Time
	sigChan      chan os.Signal
}

func NewSpinner(message string) *Spinner {
	s := &Spinner{
		parts: []string{
			"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
		},
		started: time.Now(),
		sigChan: make(chan os.Signal, 1),
	}
	s.SetMessage(message)

	// Set up signal handling
	signal.Notify(s.sigChan, syscall.SIGINT, syscall.SIGTERM)

	go s.start()
	return s
}

func (s *Spinner) SetMessage(message string) {
	s.message.Store(message)
}

func (s *Spinner) String() string {
	var sb strings.Builder
	if s.stopped.IsZero() {
		spinner := s.parts[s.value]
		sb.WriteString(spinner)
		sb.WriteString(" ")
	}
	if message, ok := s.message.Load().(string); ok && len(message) > 0 {
		message := strings.TrimSpace(message)
		if s.messageWidth > 0 && len(message) > s.messageWidth {
			message = message[:s.messageWidth]
		}
		fmt.Fprintf(&sb, "%s", message)
		if padding := s.messageWidth - sb.Len(); padding > 0 {
			sb.WriteString(strings.Repeat(" ", padding))
		}
	}
	return sb.String()
}

func (s *Spinner) start() {
	s.ticker = time.NewTicker(100 * time.Millisecond)
	fmt.Print("\r") // Start at beginning of line

	for {
		select {
		case <-s.ticker.C:
			s.value = (s.value + 1) % len(s.parts)
			if !s.stopped.IsZero() {
				return
			}
			fmt.Print("\r\033[2K") // Clear entire line
			fmt.Print(s.String())
		case sig := <-s.sigChan:
			s.Stop()
			s.Clear()
			// Reset signal handling and re-send the signal
			signal.Stop(s.sigChan)
			p, err := os.FindProcess(os.Getpid())
			if err == nil {
				p.Signal(sig)
			}
			return
		}
	}
}

func (s *Spinner) Stop() {
	if s.stopped.IsZero() {
		s.stopped = time.Now()
		if s.ticker != nil {
			s.ticker.Stop()
		}
	}
}

// Clear clears the current line using ANSI escape sequences
func (s *Spinner) Clear() {
	fmt.Print("\r\033[2K") // \033[2K clears the entire line
}
