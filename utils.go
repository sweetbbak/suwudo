package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

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

// COLOR UTILS
func formatPrompt(s, usrname string) string {
	s = replaceColor(s)
	s = strings.ReplaceAll(s, "{usr}", usrname)
	s = strings.ReplaceAll(s, "{user}", usrname)
	s = strings.ReplaceAll(s, "{USER}", usrname)
	return fmt.Sprintf(s, usrname)
}

type Hex string
type Ansi string
type RGB struct {
	R int
	G int
	B int
}

func escapeStr(s string) (string, error) {
	if len(s) < 1 {
		return "", nil
	}

	s = strings.Split(s, "\\c")[0]
	s = strings.Replace(s, "\\0", "\\", -1)
	s = fmt.Sprintf("\"%s\"", s)

	_, err := fmt.Sscanf(s, "%q", &s)
	if err != nil {
		return "", nil
	}

	return s, nil
}

func HextoRGB(hex Hex) RGB {
	if hex[0:1] == "#" {
		hex = hex[1:]
	}

	r := string(hex)[0:2]
	g := string(hex)[2:4]
	b := string(hex)[4:6]

	R, _ := strconv.ParseInt(r, 16, 0)
	G, _ := strconv.ParseInt(g, 16, 0)
	B, _ := strconv.ParseInt(b, 16, 0)

	return RGB{int(R), int(G), int(B)}

}

func HextoAnsi(hex Hex) Ansi {
	rgb := HextoRGB(hex)
	str := "\x1b[38;2;" + strconv.FormatInt(int64(rgb.R), 10) + ";" + strconv.FormatInt(int64(rgb.G), 10) + ";" + strconv.FormatInt(int64(rgb.B), 10) + "m"
	return Ansi(str)
}

func replaceColor(line string) string {
	r := regexp.MustCompile(`\{#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6})\}`)

	line = strings.ReplaceAll(line, "{clr}", "\x1b[0m")
	line = strings.ReplaceAll(line, "{clear}", "\x1b[0m")
	line = strings.ReplaceAll(line, "{CLEAR}", "\x1b[0m")

	if r.Match([]byte(line)) {
		line = r.ReplaceAllStringFunc(line, func(line string) string {
			str := line
			// str = strings.ReplaceAll(str, "{clr}", "\x1b[0m")
			// str = strings.ReplaceAll(str, "{clr}", "")
			// str = strings.ReplaceAll(str, "{clear}", "")
			str = strings.ReplaceAll(str, "{", "")
			str = strings.ReplaceAll(str, "}", "")

			hex := HextoAnsi(Hex(str))
			return string(hex)
		})
		return line
	}
	return line
}
