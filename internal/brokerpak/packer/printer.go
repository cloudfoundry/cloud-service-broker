package packer

import (
	"fmt"
	"time"
)

const layout = "2006/01/02 15:04:05"

// Println prints a message with the current date and time.
func Println(v ...any) {
	timestamp := time.Now().Format(layout)
	fmt.Println(append([]any{timestamp}, v...)...)
}

// Printf prints a formatted message with the current date and time.
func Printf(format string, v ...any) {
	timestamp := time.Now().Format(layout)
	fmt.Printf("%s "+format, append([]any{timestamp}, v...)...)
}
