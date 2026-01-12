// Package process manages child process tracking and termination.
package process

import (
	"os"
	"sync"
	"syscall"
)

//nolint:gochecknoglobals // Global state needed for signal handler coordination
var (
	activeChildren []*os.Process
	childrenMutex  sync.Mutex
)

// Register registers a child process to be killed on signal.
func Register(proc *os.Process) {
	childrenMutex.Lock()
	defer childrenMutex.Unlock()
	activeChildren = append(activeChildren, proc)
}

// Unregister removes a child process from the active list.
func Unregister(proc *os.Process) {
	childrenMutex.Lock()
	defer childrenMutex.Unlock()
	for i, p := range activeChildren {
		if p == proc {
			activeChildren = append(activeChildren[:i], activeChildren[i+1:]...)
			break
		}
	}
}

// KillAll kills all registered child processes and their process groups.
// This ensures that not only direct children are killed, but also any
// subprocesses spawned by those children (e.g., agent spawns a subprocess).
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
