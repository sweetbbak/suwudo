package sudoers

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/jessevdk/go-flags"
)

// PERMS for GROUPS or USERS
type Perms struct {
	AsUsers     []string `short:"u" long:"user" description:"run the command as the specified user"`
	AsGroups    []string `short:"g" long:"group" description:"run the command as the specified group"`
	SetEnv      []string `short:"e" long:"env" description:"set environment variables for command ex: (--env USER=suwu)"`
	Commands    []string `short:"c" long:"command" description:"allowed commands"`
	Prompt      string   `short:"p" long:"prompt" description:"use a custom prompt, use {USER} as a placeholder for the user and {#FFFFFF} for color"`
	PreserveEnv bool     `short:"E" long:"preserve-env" description:"preserve the calling users environment variables"`
	Shell       bool     `short:"s" long:"shell" description:"allow the use of a login shell"`
}

const (
	DefaultDir     = "/etc/suwu"
	ConfigFile     = "suwu.conf"
	ConfigPath     = "/etc/suwu/suwu.conf"
	DEFAULT_CONFIG = "%wheel --user=ALL --group=ALL --command=ALL --preserve-env"
)

/*
Is User allowed to do XYZ action or is USER in GROUP that is allowed to do XYZ action
*/

type User struct {
	*user.User
	Perms  *Perms
	Groups []Group
}

type Group struct {
	*user.Group
	Perms *Perms
}

func NewConf(u *user.User) (*User, error) {
	f, err := os.Open(ConfigPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	usr := &User{}
	usr.User = u

	groups, err := getUserGroups(u)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*user.Group)
	for _, g := range groups {
		m[g.Name] = g
	}

	var lineNum int
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		lineNum++

		if strings.HasPrefix(line, "#") {
			// fmt.Printf("line: [%d] IS COMMENT\n", lineNum)
			continue
		}

		// Implement additional directories
		if strings.HasPrefix(line, "@") {
			continue
		}

		// IS a GROUP
		if strings.HasPrefix(line, "%") {
			split := strings.Split(line, " ")
			gname := strings.TrimPrefix(split[0], "%")

			if gname == "" {
				return nil, fmt.Errorf("Error: group name is empty on line %d", lineNum)
			}

			// fmt.Printf("line: [%d] IS GROUP %s\n", lineNum, gname)

			if isInGroup(groups, gname) {
				// fmt.Println("USER is IN group")

				g := Group{}
				split := strings.Split(line, " ")

				if len(split) < 2 {
					return nil, fmt.Errorf("Error parsing line %d - unable to split line on whitespaces", lineNum)
				}

				var ok bool
				g.Group, ok = m[gname]

				if !ok {
					return nil, fmt.Errorf("Error parsing group name on line %d for group %s", lineNum, gname)
				}

				g.Perms, err = parseLine(line)
				if err != nil {
					return nil, err
				}

				usr.Groups = append(usr.Groups, g)

			} else {
				continue
			}
		}

		// empty lines are skipped
		if strings.TrimSpace(line) == "" {
			continue
		}

		// is a USER
		if strings.Split(line, " ")[0] == u.Username {
			usr.Perms, err = parseLine(line)
			if err != nil {
				return nil, err
			}
		}
	}

	return usr, nil
}

// is groupname STRING in array of *user.Group
func isInGroup(groups []*user.Group, gname string) bool {
	for _, g := range groups {
		if g.Name == gname {
			return true
		}
	}
	return false
}

func getUserGroups(u *user.User) ([]*user.Group, error) {
	g, err := u.GroupIds()
	if err != nil {
		return nil, err
	}

	var groups []*user.Group

	for _, gg := range g {
		grp, err := user.LookupGroupId(gg)
		if err != nil {
			return nil, err
		}
		groups = append(groups, grp)
	}
	return groups, nil
}

func parseLine(line string) (*Perms, error) {
	c := Perms{}
	parser := flags.NewParser(&c, flags.Default)
	args := strings.Split(line, " ")
	_, err := parser.ParseArgs(args)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func getConfig() {
}

// UTILS
func createConfig() error {
	_, err := os.Stat(DefaultDir)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		err = os.MkdirAll(DefaultDir, 0o755)
		if err != nil {
			return err
		}
	}

	path := filepath.Join(DefaultDir, ConfigFile)
	_, err = os.Stat(path)
	if err != nil && errors.Is(err, fs.ErrNotExist) {
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = f.WriteString(DEFAULT_CONFIG)
		if err != nil {
			return err
		}
	}
	return nil
}
