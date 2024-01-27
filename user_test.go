package main

import (
	"os/user"
	"strings"
	"testing"
)

func TestNewUser(t *testing.T) {
	usr, err := NewUser()
	if err != nil {
		t.Fatal(err)
	}

	u, err := user.Current()

	if u.Username != usr.User.Username {
		t.Fatalf("Usernames do not match")
	}
}

func TestGetTargetShell(t *testing.T) {
	usr, err := NewUser()
	if err != nil {
		t.Fatal(err)
	}

	// exec as self
	usr.asUser = usr.User

	shell, err := usr.GetTargetShell()
	if err != nil {
		t.Fatal(err)
	}

	passwdLine, err := etcPasswd(usr.User.Username, "tests/passwd")
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(passwdLine, shell) {
		t.Fail()
	}
}

// wont work needs root
// func TestExec(t *testing.T) {
// 	usr, err := NewUser()
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	usr.hasAuth = true
// 	err = usr.Exec([]string{"whoami"})
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// }
