//go:build windows

package process

// KillAll kills all registered child processes on Windows.
// On Windows, we use Process.Kill() which terminates the process.
func KillAll() {
	childrenMutex.Lock()
	defer childrenMutex.Unlock()

	for _, proc := range activeChildren {
		if proc != nil {
			// Kill the process - on Windows, this terminates the process and its children
			// Ignore errors as process might already be dead
			_ = proc.Kill()
		}
	}
	activeChildren = nil
}
