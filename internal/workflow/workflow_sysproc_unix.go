//go:build darwin || linux

package workflow

import "syscall"

// setupProcessGroup configures platform-specific process group settings for Unix systems.
func setupProcessGroup() *syscall.SysProcAttr {
	//nolint:exhaustruct // Only Setpgid is needed; other fields intentionally use zero values
	return &syscall.SysProcAttr{
		Setpgid: true, // Create new process group for proper signal handling
	}
}
