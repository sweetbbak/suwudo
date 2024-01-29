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

	Debug("NewUser: User set as: %s - Group %s\n", usr.User.Username, usr.Group.Name)

	t, ok := os.LookupEnv("TERM")
	if ok {
		usr.Env = append(usr.Env, fmt.Sprintf("TERM=%s", t))
	}

	usr.Shell = "/bin/sh"
	usr.Execute.PreserveEnv = false
	usr.Fork = false
	usr.hasAuth = false

	return usr, nil
}

func (u *User) Authorize(prompt string) error {
	pass, err := Credentials(prompt)
	if err != nil {
		return fmt.Errorf("Unable to authenticate user [%v]: %w", u.asUser.Username, err)
	}

	Debug("Verifying password for user: %s\n", u.User.Username)

	// default is /etc/shadow
	passed, err := PasswordVerify(pass, u, "")
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
	Debug("User lookup for target resolved to: %s\n", asuser)

	// get uint32 of UID for setting UID in syscall.SysProcAttr
	i, err := strconv.Atoi(asuser.Uid)
	if err != nil {
		return err
	}

	u.asUID = uint32(i)
	Debug("User lookup for target UID resolved to: %v\n", i)

	return nil
}

func (u *User) AsGroup(groupName string) error {
	group, err := user.LookupGroup(groupName)
	if err != nil {
		return err
	}

	u.asGroup = group
	Debug("User lookup for target resolved to: %s\n", group)

	i, err := strconv.Atoi(group.Gid)
	if err != nil {
		return err
	}

	u.asGID = uint32(i)
	Debug("User lookup for target UID resolved to: %v\n", i)

	return nil
}

func (u *User) GetTargetShell() (string, error) {
	// sanity check
	if u.asUser.Username == "" {
		return "", fmt.Errorf("Error getting target users default shell, internal error finding user")
	}

	passent, err := etcPasswd(u.asUser.Username, "/etc/passwd")
	if err != nil {
		return "", err
	}

	fields := strings.Split(passent, ":")
	shell := fields[len(fields)-1]

	Debug("User default shell: %v\n", shell)

	if shell != "" {
		return shell, nil
	} else {
		return "", fmt.Errorf("unable to find users shell")
	}
}

func (u *User) Exec(args []string) error {
	return Run(args, u)
}

func (u *User) ExecShell() error {
	return Run([]string{u.Shell}, u)
}

func (u *User) ExecShellCmd(args []string) error {
	cmd := strings.Join(args, " ")
	return Run([]string{u.Shell, "-c", cmd}, u)
}

func (u *User) ExecShellCmdString(cmd string) error {
	return Run([]string{u.Shell, "-c", cmd}, u)
}

func (u *User) CacheCredentials() error {
	err := CacheCreds()
	if err != nil {
		return err
	}
	return nil
}
