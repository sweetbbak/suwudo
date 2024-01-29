package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"golang.org/x/term"
	"suwu/pkg/yescrypt"
)

// Prompt user for password, and return the raw password
func Credentials(prompt string) (string, error) {
	var out *os.File
	if isPiped() {
		out = os.Stderr
	} else {
		out = os.Stdout
	}

	fmt.Fprint(out, prompt)

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

func PasswordVerify(password string, usr *User, passFile string) (bool, error) {
	// open etc shadow and find the users hash - name:$6$reallylonghash:12345:0:99999:7:::
	if passFile == "" {
		passFile = "/etc/shadow"
	}

	fi, err := os.Open(passFile)
	if err != nil {
		return false, fmt.Errorf("unable to open '%s': %w", passFile, err)
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
		return false, fmt.Errorf("unable to parse file '%s'", passFile)
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

	// TODO get rid of this C dependency. Will have to probably write a library to verify yescrypt
	if strings.HasPrefix(token, "$y$") {
		return yescrypt.Verify(password, token), nil
	}

	crypt := crypt.NewFromHash(token)

	err = crypt.Verify(token, []byte(password))
	if err != nil {
		return false, fmt.Errorf("Error verifying password against password hash: %w", err)
	} else {
		return true, nil
	}
}
