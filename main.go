package main

import (
	"bufio"
	"fmt"
	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

var (
	password  string
	buffer    string
	name      string
	full_hash string
	salt      string
	token     string
	uid       string
)

func main() {
	// root user
	uid := os.Geteuid()
	err := syscall.Setuid(0)
	if err != nil {
		fmt.Println("Error setting user as root")
		os.Exit(1)
	}

	// open pass file and read the user name from it by matching the UID
	// sweet:x:1000:1000:sweet:/home/sweet:/bin/zsh - is what it looks like
	file, err := os.Open("/etc/passwd")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), fmt.Sprintf("%d", uid)) {
			name = scanner.Text()
		}
	}

	// match UID with the line and get the first field which is the users name
	splits := strings.Split(name, ":")
	name = splits[0]

	// turn off terminal echo
	STDINFILENO := 0
	raw, err := unix.IoctlGetTermios(STDINFILENO, unix.TCGETS)
	if err != nil {
		panic(err)
	}
	rawState := *raw
	rawState.Lflag &^= unix.ECHO
	rawState.Lflag &^= unix.ICANON
	err = unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, &rawState)
	if err != nil {
		panic(err)
	}

	// get password
	fmt.Printf("Enter password for user: %s\n", name)
	fmt.Scanln(&password)

	err = unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, raw)
	if err != nil {
		panic(err)
	}

	// open etc shadow and find the users hash - name:$6$reallylonghash:12345:0:99999:7:::
	fi, err := os.Open("/etc/shadow")
	if err != nil {
		fmt.Println(err)
	}

	defer fi.Close()
	scanner1 := bufio.NewScanner(fi)

	for scanner1.Scan() {
		if strings.Contains(scanner1.Text(), name) {
			token = scanner1.Text()
		}
	}

	// hash/token is the 2nd field with a ":" delimiter
	split := strings.Split(token, ":")
	token = split[1]

	// check the hash against the password with this convenient go package
	// if err returns, pass is incorrect. simple shit.
	crypt := crypt.SHA512.New()
	err = crypt.Verify(token, []byte(password))
	if err != nil {
		fmt.Println("Incorrect password")
		os.Exit(1)
	}
	cmd := strings.Join(os.Args[1:], " ")
	exitCode := System(cmd)
	os.Exit(exitCode)
}

func System(cmd string) int {
	c := exec.Command("sh", "-c", cmd)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()
	if err == nil {
		return 0
	}

	// Figure out the exit code
	if ws, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
		if ws.Exited() {
			return ws.ExitStatus()
		}

		if ws.Signaled() {
			return -int(ws.Signal())
		}
	}
	return -1
}
