package tui

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/skratchdot/open-golang/open"
)

func isWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	cmd := exec.Command("uname", "-a")
	if output, err := cmd.Output(); err == nil {
		if strings.Contains(strings.ToLower(string(output)), "microsoft") {
			return true
		}
	}
	return false
}

func openURL(url string) error {
	if isWSL() {
		return exec.Command("powershell.exe", "Start", url).Start()
	} else {
		return open.Start(url)
	}
}
