package utils

import (
	"fmt"
	"log"
	"os"
)

// Logging function that saves the log to the log.txt file
// Every error is logged to the log.txt file

func Log(format string, a ...any) {
	log.SetOutput(os.Stdout)

	filename := "bob_log.txt"

	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf(format, a...))
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		os.Exit(1)
	}
}
