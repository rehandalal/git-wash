package helpers

import (
	"fmt"

	"github.com/mgutz/ansi"
)

func PrintC(message string, style string) {
	// Prints a colored string to the terminal
	fmt.Printf("%s%s%s", ansi.ColorCode(style), message, ansi.Reset)
}

func PrintlnC(message string, style string) {
	// Prints a colored string, suffixed with a new-line, to the terminal
	PrintC(message+"\n", style)
}
