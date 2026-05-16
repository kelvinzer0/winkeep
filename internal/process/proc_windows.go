//go:build windows

package process

import (
	"os/exec"
	"strconv"
	"syscall"
)

var (
	modkernel32     = syscall.NewLazyDLL("kernel32.dll")
	procOpenProcess = modkernel32.NewProc("OpenProcess")
	procCloseHandle = modkernel32.NewProc("CloseHandle")
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
	handle, err := openProcess(0x00100000, false, uint32(pid))
	if err != nil {
		return false
	}
	closeHandle(handle)
	return true
}

func openProcess(desiredAccess uint32, inheritHandle bool, processID uint32) (uintptr, error) {
	var inherit uintptr
	if inheritHandle {
		inherit = 1
	}
	handle, _, err := procOpenProcess.Call(
		uintptr(desiredAccess),
		inherit,
		uintptr(processID),
	)
	if handle == 0 {
		return 0, err
	}
	return handle, nil
}

func closeHandle(handle uintptr) {
	procCloseHandle.Call(handle)
}
