package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func run(args, env []string, UID, GID uint32, dir string, preserveEnv, fork bool) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: UID, Gid: GID}

	if fork {
		cmd.SysProcAttr.Setsid = true
	}

	cmd.Stdin, cmd.Stderr, cmd.Stdout = os.Stdin, os.Stderr, os.Stdout

	if !preserveEnv {
		os.Clearenv()
	} else {
		cmd.Env = os.Environ()
	}

	if len(env) != 0 {
		setEnv(env)
	}

	if dir != "" {
		cmd.Dir = dir
	}

	Debug("Env vars for command %v: %v\n", args, cmd.Environ())
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Unable to run command [%s]: %w", args, err)
	}

	return nil
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
