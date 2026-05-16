//go:build windows

package process

import (
	"os"
	"os/exec"
	"strconv"
	"syscall"
)

func setDetachAttrs(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
	}
}

func killProcess(pid int) {
	kill := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(pid))
	kill.Run()
}

func isProcessAlive(pid int) bool {
	handle, err := openProcess(0x00100000, false, uint32(pid)) // PROCESS_QUERY_LIMITED_INFORMATION
	if err != nil {
		return false
	}
	closeHandle(handle)
	return true
}
