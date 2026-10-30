package plugin

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"syscall"
)

func killTree(p *os.Process) {
	command := exec.Command(filepath.Join(os.Getenv("SystemRoot"), "System32", "taskkill.exe"), "/PID", strconv.Itoa(p.Pid), "/T", "/F")
	command.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = command.Run()
	_ = p.Kill()
}
