package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
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

func cacheCreds() error {
	if !isRoot() {
		return fmt.Errorf("Must be root")
	}

	dir, err := getConfdir()
	if err != nil {
		return err
	}

	fl := path.Join(dir, lockfile)

	Lockfile(fl)
	return nil
}

func Lockfile(lock string) error {
	file, err := os.OpenFile(lock, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o755)
	if err != nil {
		return err
	}
	defer file.Close()

	return nil
}
