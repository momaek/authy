package cmd

import (
	"fmt"
)

const (
	// Black black
	Black = "\033[1;30m%s\033[0m"
	// Red red
	Red = "\033[1;31m%s\033[0m"
	// Green green
	Green = "\033[1;32m%s\033[0m"
	// Yellow yellow
	Yellow = "\033[1;33m%s\033[0m"
	// Purple purple
	Purple = "\033[1;34m%s\033[0m"
	// Magenta magenta
	Magenta = "\033[1;35m%s\033[0m"
	// Teal teal
	Teal = "\033[1;36m%s\033[0m"
	// White white
	White = "\033[1;37m%s\033[0m"
	// DebugColor debug color
	DebugColor = "\033[0;36m%s\033[0m"
)

func prettyPrintResult(outputs []AlfredOutput) {
	fmt.Printf("\n")
	for _, tk := range outputs {
		fmt.Printf("- Title: "+Green+"\n", tk.Title)
		fmt.Printf("- Code: "+Teal+" Expires in "+Red+"(s)\n\n", tk.Code, fmt.Sprint(tk.ExpireSec))
	}
	return
}
