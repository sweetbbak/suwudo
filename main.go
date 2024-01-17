package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"

	"github.com/GehirnInc/crypt"
	_ "github.com/GehirnInc/crypt/sha512_crypt"
	"github.com/jessevdk/go-flags"
	"golang.org/x/sys/unix"
)

var opts struct {
	Directory   string   `short:"D" long:"chdir" description:"run the command in the specified directory instead of cwd"`
	AsUser      string   `short:"u" long:"user" description:"run the command as the specified user"`
	AsGroup     string   `short:"g" long:"group" description:"run the command as the specified group"`
	Prompt      string   `short:"p" long:"prompt" description:"use a custom prompt"`
	SetEnv      []string `short:"e" long:"env" description:"set environment variables for command ex: (--env USER=suwu)"`
	PreserveEnv bool     `short:"E" long:"preserve-env" description:"preserve the calling users environment variables"`
	Shell       bool     `short:"s" long:"shell" description:"preserve the calling users environment variables"`
	Fork        bool     `short:"f" long:"fork" description:"fork process into the background"`
	Verbose     bool     `short:"v" long:"verbose" description:"print debugging information and verbose output"`
}

var STDINFILENO int = 0
var Debug = func(string, ...interface{}) {}

func restore(raw *unix.Termios) error {
	err := unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, raw)
	if err != nil {
		return err
	}
	return nil
}

func askpass() (string, error) {
	// turn off terminal echo
	raw, err := unix.IoctlGetTermios(STDINFILENO, unix.TCGETS)
	if err != nil {
		return "", err
	}

	rawState := *raw
	rawState.Lflag &^= unix.ECHO
	rawState.Lflag &^= unix.ICANON
	rawState.Iflag &^= unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.ISTRIP | unix.INLCR | unix.IGNCR | unix.ICRNL | unix.IXON
	rawState.Oflag &^= unix.OPOST
	rawState.Lflag &^= unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN
	rawState.Cflag &^= unix.CSIZE | unix.PARENB
	rawState.Cflag |= unix.CS8
	rawState.Cc[unix.VMIN] = 1
	rawState.Cc[unix.VTIME] = 0

	err = unix.IoctlSetTermios(STDINFILENO, unix.TCSETS, &rawState)
	if err != nil {
		return "", err
	}

	var password string
	var prompt string

	if opts.Prompt != "" {
		// TODO add checking to prompt so no escape sequences are absurd length are passed
		// needs fuzzing
		prompt = opts.Prompt
	} else {
		prompt = "\x1b[38;2;111;111;111m│ \x1b[0m\x1b[38;2;124;120;254mpassword for \x1b[38;2;245;127;224m\x1b[3m%s\x1b[0m \x1b[38;2;124;120;254m>\x1b[0m "
	}

	// not super important its just for the prompt
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}

	fmt.Fprintf(os.Stderr, "\x1b[2K")
	fmt.Fprintf(os.Stderr, "\x1b[0G") // clear line
	fmt.Fprintf(os.Stderr, prompt, user)
	// fmt.Fscanf(os.Stdout, "%s", &password)
	fmt.Fscanf(os.Stdin, "%s", &password)

	// erase line
	fmt.Fprintf(os.Stderr, "\x1b[2K")
	fmt.Fprintf(os.Stderr, "\x1b[0G")

	if err := restore(raw); err != nil {
		return "", fmt.Errorf("Error restoring terminal state: %v", err)
	}

	fmt.Fprintf(os.Stderr, "\x1b[2K")

	return password, nil
}

func get_user() string {
	var name string
	uid := os.Geteuid()
	fmt.Println(uid)
	// open pass file and read the user name from it by matching the UID
	// sweet:x:1000:1000:sweet:/home/sweet:/bin/zsh - is what it looks like
	file, err := os.Open("/etc/passwd")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), fmt.Sprintf("%d", uid)) {
			name = scanner.Text()
		}
	}

	// match UID with the line and get the first field which is the users name
	splits := strings.Split(name, ":")
	name = splits[0]
	return name
}

func verify_pass(password string, uid int) (bool, error) {
	var token string
	UID := fmt.Sprintf("%d", uid)

	name, err := user.LookupId(UID)
	if err != nil {
		return false, err
	}

	// open etc shadow and find the users hash - name:$6$reallylonghash:12345:0:99999:7:::
	fi, err := os.Open("/etc/shadow")
	if err != nil {
		return false, fmt.Errorf("unable to open /etc/shadow: %v", err)
	}
	defer fi.Close()

	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()
		username := strings.SplitN(line, ":", 2)[0]

		if username == name.Username || username == name.Name {
			token = line
		} else {
			continue
		}
	}

	if token == "" {
		return false, fmt.Errorf("unable to parse shadow file in /etc/shadow")
	}

	// hash/token is the 2nd field with a ":" delimiter
	split := strings.Split(token, ":")
	token = split[1]

	if token[0] != '$' {
		return false, fmt.Errorf("error, password hash appears to be incorrect") // needs better message
	}

	// check the hash against the password with this convenient go package
	// if err returns, pass is incorrect. simple shit.
	crypt := crypt.SHA512.New()
	err = crypt.Verify(token, []byte(password))
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func setEnv(args []string) {
	for _, v := range args {
		if strings.IndexByte(v, '=') > 0 {
			kv := strings.SplitN(v, "=", 2) // split key/value pair in 2
			if err := os.Setenv(kv[0], kv[1]); err != nil {
				log.Println(err)
			}
		}
	}
}

func Suwu(args []string) error {
	// learned this the hard way to get the UID of the OG user BEFORE
	// calling Setuid() lol. Otherwise it will return 0 for root
	userID := syscall.Getuid()
	if userID == -1 {
		return fmt.Errorf("error getting user ID")
	}

	groupID := syscall.Getgid()
	if groupID == -1 {
		return fmt.Errorf("error getting group ID")
	}

	oguser, err := user.Current()
	if err != nil {
		return err
	}

	// get effective user ID and set to root user
	err = syscall.Setuid(0)
	if err != nil {
		return fmt.Errorf("Error setting user as root, ensure binary has SETUID permissions set")
	}

	pass, err := askpass()
	if err != nil {
		return err
	}

	passed, err := verify_pass(pass, userID)
	if err != nil {
		return err
	}

	if !passed {
		return fmt.Errorf("Incorrect password")
	}

	if len(args) == 0 && !opts.Shell {
		return nil
	}

	u, err := NewUser()
	if err != nil {
		return err
	}
	u.Authorize()
	u.AsUser(opts.AsUser)
	u.AsGroup(opts.AsGroup)
	shell, err := u.GetTargetShell()
	if err != nil {
		return err
	}

	if opts.Shell {
		return u.Exec([]string{shell})
	}

	return runAsUser(args, oguser)
}

func runAsUser(args []string, oguser *user.User) error {
	var userid, groupid uint32
	var usern *user.User
	var group *user.Group
	var err error

	if opts.AsGroup != "" {
		group, err = user.LookupGroup(opts.AsGroup)
		if err != nil {
			return err
		}

		i, err := strconv.Atoi(group.Gid)
		if err != nil {
			return err
		}

		groupid = uint32(i)
	} else {
		groupid = 0
	}

	if opts.AsUser != "" {
		usern, err = user.Lookup(opts.AsUser)
		if err != nil {
			return err
		}

		i, err := strconv.Atoi(usern.Uid)
		if err != nil {
			return err
		}

		userid = uint32(i)
	} else {
		usern, err = user.Lookup("root")
		if err != nil {
			return err
		}

		i, err := strconv.Atoi(usern.Uid)
		if err != nil {
			return err
		}

		userid = uint32(i) // should always be 0 but im trying this out
	}

	passent, err := etcPasswd(usern.Username)
	if err != nil {
		return err
	}

	if strings.Contains(passent, "nologin") {
		return fmt.Errorf("This account is currently not available")
	}

	fields := strings.Split(passent, ":")
	shell := fields[len(fields)-1]

	Debug("exec as user: UID [%v] GID [%v]\n", userid, groupid)

	if opts.Shell {
		return run([]string{shell}, userid, groupid) // 0, 0 is ignored here
	} else {
		return run(args, userid, groupid)
	}
}

// return users etc passwd entry
func etcPasswd(usern string) (string, error) {
	fi, err := os.Open("/etc/passwd")
	if err != nil {
		return "", err
	}
	defer fi.Close()

	if usern == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	scanner := bufio.NewScanner(fi)
	for scanner.Scan() {
		line := scanner.Text()

		username := strings.SplitN(line, ":", 2)[0]
		if username == usern {
			return line, nil
		} else {
			continue
		}
	}
	return "", fmt.Errorf("unable to find user")
}

func run(args []string, UID, GID uint32) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.SysProcAttr = &syscall.SysProcAttr{}

	// cmd.SysProcAttr.Credential.Uid = UID
	// cmd.SysProcAttr.Credential.Gid = GID
	cmd.SysProcAttr.Credential = &syscall.Credential{Uid: UID, Gid: GID}

	if opts.Fork {
		cmd.SysProcAttr.Setsid = true
	}

	cmd.Stdin, cmd.Stderr, cmd.Stdout = os.Stdin, os.Stderr, os.Stdout

	if !opts.PreserveEnv {
		os.Clearenv()
	} else {
		cmd.Env = os.Environ()
	}

	if len(opts.SetEnv) != 0 {
		setEnv(opts.SetEnv)
	}

	if opts.Directory != "" {
		err := os.Chdir(opts.Directory)
		if err != nil {
			return err
		}
	}

	Debug("Env vars for command %v: %v\n", args, cmd.Environ())

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func main() {
	args, err := flags.Parse(&opts)
	if flags.WroteHelp(err) {
		os.Exit(0)
	}
	if err != nil {
		log.Fatal(err)
	}

	if opts.Verbose {
		Debug = log.Printf
	}

	if err := Suwu(args); err != nil {
		log.Fatal(err)
	}
}

func system(cmd string) int {
	c := exec.Command("sh", "-c", cmd)
	c.Env = os.Environ()
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	err := c.Run()
	if err == nil {
		return 0
	}

	// Figure out the exit code
	if ws, ok := c.ProcessState.Sys().(syscall.WaitStatus); ok {
		if ws.Exited() {
			return ws.ExitStatus()
		}

		if ws.Signaled() {
			return -int(ws.Signal())
		}
	}
	return -1
}
