package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jessevdk/go-flags"
)

var opts struct {
	Directory   string   `short:"D" long:"chdir" description:"run the command in the specified directory instead of cwd"`
	AsUser      string   `short:"u" long:"user" description:"run the command as the specified user"`
	AsGroup     string   `short:"g" long:"group" description:"run the command as the specified group"`
	Prompt      string   `short:"p" long:"prompt" description:"use a custom prompt, use {USER} as a placeholder for the user and {#FFFFFF} for color"`
	SetEnv      []string `short:"e" long:"env" description:"set environment variables for command ex: (--env USER=suwu)"`
	PreserveEnv bool     `short:"E" long:"preserve-env" description:"preserve the calling users environment variables"`
	Shell       bool     `short:"s" long:"shell" description:"use the default shell for the user we are executing as"`
	ExecShell   bool     `short:"S" long:"exec" description:"execute command using SHELL"`
	Login       bool     `short:"l" long:"login" description:"use shell as user and source all of their environment variables"`
	Fork        bool     `short:"f" long:"fork" description:"fork process into the background"`
	Verbose     bool     `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var Debug = func(string, ...interface{}) {}

func Suwu(args []string) error {
	usr, err := NewUser()
	if err != nil {
		return err
	}

	var prompt string
	if opts.Prompt != "" {
		prompt = opts.Prompt
	} else {
		prompt = "\x1b[38;2;111;111;111m│ \x1b[0m\x1b[38;2;124;120;254mpassword for \x1b[38;2;245;127;224m\x1b[3m%s\x1b[0m \x1b[38;2;124;120;254m>\x1b[0m "
		prompt = fmt.Sprintf(prompt, usr.User.Username)
	}

	err = usr.Authorize(prompt)
	if err != nil {
		return err
	}

	// redundant check for authorization
	if !usr.hasAuth {
		return fmt.Errorf("User not authorized")
	}

	// set user and group if set
	if opts.AsUser != "" {
		Debug("main: setting user: %s\n", opts.AsUser)
		err := usr.AsUser(opts.AsUser)
		if err != nil {
			return err
		}
	} else {
		err := usr.AsUser("root")
		if err != nil {
			return err
		}
	}

	if opts.AsGroup != "" {
		err := usr.AsGroup(opts.AsGroup)
		if err != nil {
			return err
		}
	}

	var shell string
	shell, err = usr.GetTargetShell()
	if err != nil {
		shell = "/bin/sh"
	}

	usr.Shell = shell

	// interactive shell
	if opts.Shell {
		return usr.ExecShell(args)
	}

	// exec CMD with shell
	if opts.ExecShell {
		return usr.ExecShellCmd(args)
	}

	return usr.Exec(args)
}

func main() {
	parser := flags.NewParser(&opts, flags.Default)
	args, err := parser.Parse()
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if len(args) < 1 && !opts.Shell {
		parser.WriteHelp(os.Stderr)
		os.Exit(0)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Suwu(args); err != nil {
		log.Fatal(err)
	}
}
