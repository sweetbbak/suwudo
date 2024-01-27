package main

import (
	"os/exec"
	"os/user"
	"strings"
	"testing"
)

func TestCredentials(t *testing.T) {
	subproc := exec.Command("./suwu", "echo", "test")
	input := "sweet\n"
	subproc.Stdin = strings.NewReader(input)
	output, _ := subproc.Output()

	if strings.Contains(string(output), "test") {
		t.Errorf("Wanted: %v, Got: %v", input, string(output))
	}
	subproc.Wait()
}

func TestPasswordVerify(t *testing.T) {
	usr := &User{}
	u := &user.User{
		Uid:      "1000",
		Gid:      "1000",
		Username: "test",
	}

	usr.User = u

	subproc := exec.Command("./tests/gen_shadow.sh")
	subproc.Run()
	subproc.Wait()

	auth, err := PasswordVerify("hunter2", usr, "./tests/shadow")
	if err != nil {
		t.Fatal(err)
	}

	if !auth {
		t.Fatal("password wasnt verified")
	}
}
