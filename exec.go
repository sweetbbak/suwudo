package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func Run(args []string, u *User) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: u.asUID, Gid: u.asGID}

	if u.Fork {
		cmd.SysProcAttr.Setsid = true
	}

	cmd.Stdin, cmd.Stderr, cmd.Stdout = os.Stdin, os.Stderr, os.Stdout

	if u.asUser.HomeDir != "" {
		cmd.Env = append(cmd.Env, fmt.Sprintf("HOME=%s", u.asUser.HomeDir))
	}

	// mutate cmd env vars to preserve specific variables
	preserveEnvVars(cmd)

	if !u.PreserveEnv {
		os.Clearenv()
	} else {
		cmd.Env = os.Environ()
	}

	if len(u.Env) != 0 {
		setEnvCmd(u.Env, cmd)
	}

	if u.Dir != "" {
		cmd.Dir = u.Dir
	}

	Debug("user: %v\n", u)

	Debug("Env vars for command %v: %v\n", args, cmd.Environ())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Unable to run command %s: %w", args, err)
	}

	return nil
}

func preserveEnvVars(cmd *exec.Cmd) {
	envvars := []string{"TERM", "PATH", "EDITOR", "TZ", "LANG", "XDG_CURRENT_DESKTOP", "DISPLAY", "COLORTERM", "BROWSER"}
	for _, k := range envvars {
		v, ok := os.LookupEnv(k)
		if ok {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
		}
	}
}

// interactive shell
func runShell(cmd, shell string) error {
	c := exec.Command(shell)
	c.Env = os.Environ()
	c.Stdin, c.Stderr, c.Stdout = os.Stdin, os.Stderr, os.Stdout

	if err := c.Run(); err != nil {
		return fmt.Errorf("Unable to run command [%s]: %w", c.String(), err)
	}

	return nil
}

// sh -c 'CMD ARGS'
func runShellCmd(cmd, shell string) error {
	c := exec.Command(shell, "-c", cmd)
	c.Env = os.Environ()
	c.Stdin, c.Stderr, c.Stdout = os.Stdin, os.Stderr, os.Stdout

	if err := c.Run(); err != nil {
		return fmt.Errorf("Unable to run command [%s]: %w", c.String(), err)
	}

	return nil
}
