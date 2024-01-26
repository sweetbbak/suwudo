package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// Prompt user for password, and return the raw password
func Credentials(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("unable to read password: %w", err)
	}

	password := string(bytePassword)
	if len(password) > 1 && strings.HasSuffix("\n", password) {
		password = strings.TrimSuffix("\n", password)
	}

	return strings.TrimSpace(password), nil
}
