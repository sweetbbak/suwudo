package main

import (
	"fmt"
	"os"
	"path"
	"syscall"
	"time"
)

// TODO create proper root-only config directory and user specific cache
var confDir = "/etc/suwu"
var lockfile = "suwu.lock"

func isRoot() bool {
	euid := syscall.Geteuid()
	return euid == 0
}

func getConfdir() (string, error) {
	_, err := os.Stat(confDir)
	if err == nil {
		return confDir, nil
	}
	return "", fmt.Errorf("could not find a suitable directory for lockfile")
}

func CacheCreds() error {
	dir, err := getConfdir()
	if err != nil {
		return err
	}

	fl := path.Join(dir, lockfile)

	err = Lockfile(fl)
	if err != nil {
		return err
	}

	return nil
}

func Lockfile(lock string) error {
	file, err := os.OpenFile(lock, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	now := time.Now().Add(time.Minute * 15)
	_, err = file.WriteString(now.Format(time.Layout))
	if err != nil {
		return err
	}

	return nil
}

// checks lock file for 10 minute credential timeout
func IsAuthorizedCache() (bool, error) {
	lfile := path.Join(confDir, lockfile)
	_, err := os.Stat(lfile)
	if err != nil {
		return false, err
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
	Debug("time now: %v\nexpiry: %v -- Is before timeout?: %v\n", t, now, now.Before(t))

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
