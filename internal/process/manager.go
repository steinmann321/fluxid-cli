// Package process manages child process tracking and termination.
package process

import (
	"os"
	"sync"
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
// Platform-specific implementation in manager_unix.go and manager_windows.go
