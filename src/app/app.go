package app

import (
	"io"
	"log"
)

// SetupLogger configures the logger to write to the provided writer and disables timestamps.
func SetupLogger(w io.Writer) {
	log.SetOutput(w)
	log.SetFlags(0)
}

// Hello returns the greeting used by the application.
func Hello() string {
	return "hello world"
}

// PrintHello writes the greeting using the configured logger.
func PrintHello() {
	log.Println(Hello())
}
