package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"time"
)

var configDirs = []string{"/var/lock", "/etc", "/var", "/run"}
var lockfile = "suwu.lock"

func isRoot() bool {
	cuser, _ := user.Current()
	return cuser.Uid == "0"
}

func getConfdir() (string, error) {
	for _, dir := range configDirs {
		_, err := os.Stat(dir)
		if err == nil {
			return dir, nil
		}
	}
	return "", fmt.Errorf("could not find a suitable directory for lockfile")
}

func CacheCreds() error {
	// if !isRoot() {
	// 	return fmt.Errorf("Must be root")
	// }

	dir, err := getConfdir()
	if err != nil {
		return err
	}

	fl := path.Join(dir, lockfile)

	Lockfile(fl)
	return nil
}

func Lockfile(lock string) error {
	file, err := os.OpenFile(lock, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	defer file.Close()

	now := time.Now().Add(time.Second * 15)
	_, err = file.WriteString(now.Format(time.Layout))
	if err != nil {
		return err
	}

	return nil
}

// checks lock file for 10 minute credential timeout
func IsAuthorizedCache() (bool, error) {
	var lfile string
	for _, dir := range configDirs {
		file := path.Join(dir, lockfile)
		_, err := os.Stat(file)
		if err == nil {
			lfile = file
			break
		}
	}

	if lfile == "" {
		return false, fmt.Errorf("lock file not found")
	}

	Debug("Found lock file: %s", lfile)

	b, err := os.ReadFile(lfile)
	if err != nil {
		return false, err
	}

	t, err := time.Parse(time.Layout, string(b))
	if err != nil {
		return false, err
	}

	now := time.Now()
	Debug("time now: %v\nexpiry: %v -- Is before?: %v\n", t, now, now.Before(t))

	if now.Before(t) || now.Equal(t) {
		Debug("Is before limit")
		return true, nil
	}

	if now.After(t) {
		err := os.Remove(lfile)
		if err != nil {
			Debug("error removing lock file: %w\n", err)
		}
		return false, nil
	}

	return false, fmt.Errorf("Error parsing credentials")
}
