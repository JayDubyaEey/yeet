package ui

import (
	"fmt"
	"github.com/fatih/color"
	"os"
)

var (
	noColor bool
	verbose bool

	infoPrefix    = "ℹ "
	warnPrefix    = "⚠ "
	successPrefix = "✔ "
	errorPrefix   = "✖ "

	infoColor    = color.New(color.FgCyan)
	warnColor    = color.New(color.FgYellow)
	successColor = color.New(color.FgGreen)
	errorColor   = color.New(color.FgRed)
)

// Setup configures the UI package
func Setup(disableColor, verboseMode bool) {
	noColor = disableColor || os.Getenv("NO_COLOR") != ""
	verbose = verboseMode

	if noColor {
		color.NoColor = true
		// Use ASCII fallbacks
		infoPrefix = "[INFO] "
		warnPrefix = "[WARN] "
		successPrefix = "[OK] "
		errorPrefix = "[ERROR] "
	}
}

// Info prints an info message
func Info(format string, args ...interface{}) {
	if !verbose {
		return
	}
	msg := fmt.Sprintf(format, args...)
	infoColor.Printf("%s%s\n", infoPrefix, msg)
}

// Warn prints a warning message
func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	warnColor.Printf("%s%s\n", warnPrefix, msg)
}

// Success prints a success message
func Success(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	successColor.Printf("%s%s\n", successPrefix, msg)
}

// Error prints an error message
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	errorColor.Fprintf(os.Stderr, "%s%s\n", errorPrefix, msg)
}
