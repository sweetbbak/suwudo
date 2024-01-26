package main

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
)

// user is the user - group is optional
// asUID and GID are what we are changing the users ID to
type User struct {
	User    *user.User
	Group   *user.Group
	asUser  *user.User
	asGroup *user.Group
	asUID   uint32
	asGID   uint32
	hasAuth bool
	Execute
}

type Execute struct {
	Env         []string
	Dir         string
	Shell       string
	PreserveEnv bool
	Fork        bool
}

func NewUser() (*User, error) {
	usr := &User{}
	var err error

	usr.User, err = user.Current()
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %v", err)
	}

	usr.Group, err = user.LookupGroup(usr.User.Username) // default should be a group of the same name as the user
	if err != nil {
		usr.asGID = 0 // default to root user
	}

	dir, err := os.Getwd()
	if err == nil {
		usr.Execute.Dir = dir
	}

	// defaults will be overridden
	usr.Execute.Env = os.Environ()
	usr.Shell = "/bin/bash"
	usr.Execute.PreserveEnv = false
	usr.Fork = false

	return usr, nil
}

func (u *User) Authorize(prompt string) error {
	pass, err := Credentials(prompt)
	if err != nil {
		return fmt.Errorf("Unable to authenticate user [%v]: %w", u.asUser.Username, err)
	}

	passed, err := PasswordVerify(pass, u)
	if err != nil {
		return err
	}

	if passed {
		u.hasAuth = true
		return nil
	} else {
		u.hasAuth = false
		return fmt.Errorf("Unable to authenticate user [%v]: %w", u.asUser.Username, err)
	}
}

func (u *User) AsUser(username string) error {
	asuser, err := user.Lookup(username)
	if err != nil {
		return err
	}

	u.asUser = asuser

	// get uint32 of UID for setting UID in syscall.SysProcAttr
	i, err := strconv.Atoi(asuser.Uid)
	if err != nil {
		return err
	}

	u.asUID = uint32(i)
	return nil
}

func (u *User) AsGroup(groupName string) error {
	group, err := user.LookupGroup(groupName)
	if err != nil {
		return err
	}

	u.asGroup = group

	i, err := strconv.Atoi(group.Gid)
	if err != nil {
		return err
	}

	u.asGID = uint32(i)
	return nil
}

func (u *User) GetTargetShell() (string, error) {
	passent, err := etcPasswd(u.User.Username)
	if err != nil {
		return "", err
	}

	fields := strings.Split(passent, ":")
	shell := fields[len(fields)-1]

	if shell != "" {
		return shell, nil
	} else {
		return "", fmt.Errorf("unable to find users shell")
	}
}

func (u *User) Exec(args []string) error {
	return run(args, u.Env, u.asUID, u.asGID, u.Dir, u.PreserveEnv, u.Fork)
}

func (u *User) ExecShell(args []string) error {
	cmd := strings.Join(args, " ")
	return runShell(cmd, u.Shell)
}

func (u *User) ExecShellCmd(args []string) error {
	cmd := strings.Join(args, " ")
	return runShell(cmd, u.Shell)
}
