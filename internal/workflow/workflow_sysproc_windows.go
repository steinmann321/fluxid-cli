//go:build windows

package workflow

import "syscall"

// setupProcessGroup configures platform-specific process group settings for Windows.
func setupProcessGroup() *syscall.SysProcAttr {
	//nolint:exhaustruct // Only CreationFlags is needed; other fields intentionally use zero values
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}
