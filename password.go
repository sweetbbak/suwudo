package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"golang.org/x/term"
)

// Prompt user for password, and return the raw password
func Credentials(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("unable to read password: %w", err)
	}

	// erase line
	fmt.Fprintf(os.Stderr, "\x1b[2K")
	fmt.Fprintf(os.Stderr, "\x1b[0G")

	password := string(bytePassword)
	if len(password) > 1 && strings.HasSuffix("\n", password) {
		password = strings.TrimSuffix("\n", password)
	}

	return strings.TrimSpace(password), nil
}

func PasswordVerify(password string, usr *User) (bool, error) {
	// open etc shadow and find the users hash - name:$6$reallylonghash:12345:0:99999:7:::
	var passFile = "/etc/shadow"
	fi, err := os.Open(passFile)
	if err != nil {
		return false, fmt.Errorf("unable to open /etc/shadow: %w", err)
	}
	defer fi.Close()

	var token string
	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()
		username := strings.SplitN(line, ":", 2)[0]
		if username == usr.User.Username {
			token = line
		} else {
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("Error while reading file: %s: %w", passFile, err)
	}

	if token == "" {
		return false, fmt.Errorf("unable to parse shadow file in /etc/shadow")
	}

	// hash/token is the 2nd field with a ":" delimiter
	split := strings.Split(token, ":")
	if len(split) < 1 {
		return false, fmt.Errorf("Unable to find password hash in password file '%s'", passFile)
	} else {
		token = split[1]
	}

	if token[0] != '$' {
		return false, fmt.Errorf("Error: password hash appears to be in an incorrect format or there was an error with scanning '%s'", passFile)
	}

	crypt := crypt.SHA512.New()
	err = crypt.Verify(token, []byte(password))
	if err != nil {
		return false, fmt.Errorf("Error verifying password against password hash: %w", err)
	} else {
		return true, nil
	}
}
