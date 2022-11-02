package helpers

import (
	"fmt"

	"github.com/mgutz/ansi"
)

func PrintColorized(message string, style string) {
	// Prints a colored string to the terminal
	fmt.Printf("%s%s%s", ansi.ColorCode(style), message, ansi.Reset)
}

func PrintlnColorized(message string, style string) {
	// Prints a colored string, suffixed with a new-line, to the terminal
	PrintColorized(message+"\n", style)
}
