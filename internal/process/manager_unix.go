//go:build darwin || linux

package process

import "syscall"

// KillAll kills all registered child processes and their process groups.
// On Unix systems, we kill the entire process group using negative PID.
func KillAll() {
	childrenMutex.Lock()
	defer childrenMutex.Unlock()

	for _, proc := range activeChildren {
		if proc != nil && proc.Pid > 0 {
			// Kill the entire process group (negative PID kills the group)
			// This kills the agent AND any subprocesses it spawned
			// Ignore errors as process might already be dead
			_ = syscall.Kill(-proc.Pid, syscall.SIGKILL)
		}
	}
	activeChildren = nil
}
