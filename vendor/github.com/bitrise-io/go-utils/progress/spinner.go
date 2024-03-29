package progress

import (
	"fmt"
	"io"
	"os"
	"time"
	"unicode/utf8"
)

// Spinner ...
type Spinner struct {
	message string
	chars   []string
	delay   time.Duration
	writer  io.Writer

	active     bool
	lastOutput string
	stopChan   chan bool
}

// NewSpinner ...
func NewSpinner(message string, chars []string, delay time.Duration, writer io.Writer) Spinner {
	return Spinner{
		message: message,
		chars:   chars,
		delay:   delay,
		writer:  writer,

		active:   false,
		stopChan: make(chan bool),
	}
}

// NewDefaultSpinner ...
func NewDefaultSpinner(message string) Spinner {
	return NewDefaultSpinnerWithOutput(message, os.Stdout)
}

// NewDefaultSpinnerWithOutput ...
func NewDefaultSpinnerWithOutput(message string, output io.Writer) Spinner {
	chars := []string{"⣾", "⣽", "⣻", "⢿", "⡿", "⣟", "⣯", "⣷"}
	delay := 100 * time.Millisecond
	return NewSpinner(message, chars, delay, output)
}

func (s *Spinner) erase() {
	n := utf8.RuneCountInString(s.lastOutput)
	for _, c := range []string{"\b", " ", "\b"} {
		for i := 0; i < n; i++ {
			if _, err := fmt.Fprintf(s.writer, c); err != nil {
				fmt.Printf("failed to update progress, error: %s\n", err)
			}
		}
	}
	s.lastOutput = ""
}

// Start ...
func (s *Spinner) Start() {
	if s.active {
		return
	}
	s.active = true

	go func() {
		for {
			for i := 0; i < len(s.chars); i++ {
				select {
				case <-s.stopChan:
					return
				default:
					s.erase()

					out := fmt.Sprintf("%s %s", s.message, s.chars[i])
					if _, err := fmt.Fprint(s.writer, out); err != nil {
						fmt.Printf("failed to update progress, error: %s\n", err)
					}
					s.lastOutput = out

					time.Sleep(s.delay)
				}
			}
		}
	}()
}

// Stop ...
func (s *Spinner) Stop() {
	if s.active {
		s.active = false
		s.erase()
		s.stopChan <- true
	}
}
