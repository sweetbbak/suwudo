package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	defaultPrompt = "\x1b[38;2;111;111;111m│ \x1b[0m\x1b[38;2;124;120;254mpassword for \x1b[38;2;245;127;224m\x1b[3m%s\x1b[0m \x1b[38;2;124;120;254m>\x1b[0m "
)

// TODO: add themes
func getPrompt(uname string) string {
	if opts.Prompt != "" {
		return formatPrompt(opts.Prompt, uname)
	} else {
		if !isColorterm() {
			return formatPrompt("password for user %s: ", uname)
		}

		if DefaultPrompt == "" {
			return formatPrompt(defaultPrompt, uname)
		} else {
			return formatPrompt(DefaultPrompt, uname)
		}
	}
}

func isColorterm() bool {
	_, ok := os.LookupEnv("COLORTERM")
	return ok
}

// is stdout a pipe?
func isPiped() bool {
	f, err := os.Stdout.Stat()
	if err != nil {
		return false
	}

	return f.Mode()&os.ModeCharDevice == 0
}

// COLOR UTILS
func formatPrompt(s, usrname string) string {
	if len(s) > 100 {
		return fmt.Sprintf("prompt for %s: ", usrname)
	}

	s, _ = escapeStr(s)
	s = replaceColor(s)
	s = strings.ReplaceAll(s, "{usr}", usrname)
	s = strings.ReplaceAll(s, "{user}", usrname)
	s = strings.ReplaceAll(s, "{USER}", usrname)
	return s
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
			str = strings.ReplaceAll(str, "{", "")
			str = strings.ReplaceAll(str, "}", "")

			hex := HextoAnsi(Hex(str))
			return string(hex)
		})
		return line
	}
	return line
}
