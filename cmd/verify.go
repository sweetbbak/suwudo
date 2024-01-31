package main

import (
	"fmt"
	"os/user"
	"strings"

	"sudo/pkg/sudoers"
)

func verifyUser(su *sudoers.User) (bool, error) {
	var isAllowed bool

	for _, asuser := range su.Perms.AsUsers {
		if strings.ToLower(asuser) == "all" {
			isAllowed = true
		}

		if opts.AsUser == asuser {
			isAllowed = true
		}
	}

	if isAllowed {
		return true, nil
	} else {
		return false, fmt.Errorf("User is %v is not allowed to execute as user %v", su.Username, opts.AsUser)
	}
}

func verifyGroup(su *sudoers.User) (bool, error) {
	var isAllowed bool

	// iterate over the groups the user is in
	for _, grp := range su.Groups {
		// these groups allow executing as these users
		for _, usr := range grp.Perms.AsUsers {
			if strings.ToLower(usr) == "all" {
				isAllowed = true
			}

			if usr == opts.AsUser {
				isAllowed = true
			}
		}
	}

	if opts.AsGroup != "" {
		for _, g := range su.Perms.AsGroups {
			if g == opts.AsGroup || strings.ToLower(g) == "all" {
				isAllowed = true
			}
		}
	}

	if isAllowed {
		return true, nil
	} else {
		return false, fmt.Errorf("User is %v is not allowed to execute as user %v", su.Username, opts.AsUser)
	}
}

func commandVerify(args []string, su *sudoers.User) (bool, error) {
	if len(args) < 1 && opts.UseShell == "" {
		return false, fmt.Errorf("no command and no shell command")
	}

	var isAllowed bool
	var cmd string

	if len(args) > 0 {
		cmd = args[0]
	}

	if opts.UseShell != "" && len(args) == 0 {
		// check first part of shell command
		cmd = strings.Split(opts.UseShell, " ")[0]
	}

	for _, c := range su.Perms.Commands {
		if strings.ToLower(c) == "all" {
			isAllowed = true
		}

		if strings.TrimSpace(cmd) == strings.TrimSpace(c) {
			isAllowed = true
		}

	}

	if isAllowed {
		return true, nil
	} else {
		return false, fmt.Errorf("User is %v is not allowed to execute %v", su.Username, cmd)
	}
}

func Verify(u *user.User, args []string) error {
	su, err := sudoers.NewConf(u)
	if err != nil {
		return err
	}

	userVerify, err := verifyUser(su)
	if err != nil {
		return err
	}

	groupVerify, err := verifyGroup(su)
	if err != nil {
		return err
	}

	cmdVerify, err := commandVerify(args, su)
	if err != nil {
		return err
	}

	if len(su.Perms.SetEnv) > 0 {
		opts.SetEnv = append(opts.SetEnv, su.Perms.SetEnv...)
	}

	if opts.Shell {
		if !su.Perms.Shell {
			return fmt.Errorf("Shell session is not allowed")
		}
	}

	if !userVerify && !groupVerify {
		return fmt.Errorf("User is %v is not allowed to execute as user %v", su.Username, opts.AsUser)
	}

	if !cmdVerify {
		return fmt.Errorf("User is %v is not allowed to execute cmd %v", su.Username, args)
	}

	if su.Perms.Prompt != "" {
		DefaultPrompt = formatPrompt(su.Perms.Prompt, u.Username)
	}

	return nil
}
