package compile

import (
	"os"
	"os/exec"
	"strings"
)

const (
	ExecCmdArgsOffset = 1
)

func ExecCmd(tool string, args ...string) *exec.Cmd {
	c := exec.Command(tool, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c
}

func ExtractTrimpathFromArgs(args []string) string {
	for i, arg := range args {
		if arg == "-trimpath" && i+1 < len(args) {
			return strings.ReplaceAll(args[i+1], "=>", "")
		}
	}
	return ""
}
