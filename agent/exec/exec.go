package exec

import (
	"os/exec"
	"runtime"
	"strings"
)

// ExecuteCommand executes a command and returns the output
func ExecuteCommand(command string) (string, error) {
	var cmd *exec.Cmd

	if strings.HasPrefix(command, "ps:") {
		psCommand := strings.TrimPrefix(command, "ps:")
		if runtime.GOOS == "windows" {
			cmd = exec.Command("powershell", "-Command", psCommand)
		} else {
			cmd = exec.Command("pwsh", "-Command", psCommand)
		}
	} else {
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/c", command)
		} else {
			cmd = exec.Command("sh", "-c", command)
		}
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}

	return string(output), nil
}
