package process

import (
	"os"
	"sync"
	"testing"
)

//nolint:paralleltest // Tests share global state and must run sequentially
func TestRegisterUnregister(t *testing.T) {
	// No t.Parallel() - tests share global state and must run sequentially

	// Clear any existing children from previous tests
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		activeChildren = nil
	}()

	// Create a dummy process (we'll use the current process as a placeholder)
	proc := &os.Process{Pid: os.Getpid()}

	// Register the process
	Register(proc)

	// Verify it was registered
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 1 {
			t.Errorf("Expected 1 child process, got %d", len(activeChildren))
		}
		if activeChildren[0] != proc {
			t.Error("Registered process doesn't match")
		}
	}()

	// Unregister the process
	Unregister(proc)

	// Verify it was unregistered
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 0 {
			t.Errorf("Expected 0 child processes after unregister, got %d", len(activeChildren))
		}
	}()
}

//nolint:paralleltest // Tests share global state and must run sequentially
func TestRegisterMultiple(t *testing.T) {
	// No t.Parallel() - tests share global state and must run sequentially

	// Clear any existing children
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		activeChildren = nil
	}()

	// Create multiple dummy processes
	proc1 := &os.Process{Pid: 1000}
	proc2 := &os.Process{Pid: 2000}
	proc3 := &os.Process{Pid: 3000}

	// Register all processes
	Register(proc1)
	Register(proc2)
	Register(proc3)

	// Verify all were registered
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 3 {
			t.Errorf("Expected 3 child processes, got %d", len(activeChildren))
		}
	}()

	// Unregister middle process
	Unregister(proc2)

	// Verify correct process was removed
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 2 {
			t.Errorf("Expected 2 child processes after unregister, got %d", len(activeChildren))
		}
		found1 := false
		found3 := false
		for _, p := range activeChildren {
			if p == proc1 {
				found1 = true
			}
			if p == proc3 {
				found3 = true
			}
		}
		if !found1 || !found3 {
			t.Error("Wrong processes remain after unregister")
		}
	}()
}

//nolint:paralleltest // Tests share global state and must run sequentially
func TestUnregisterNonexistent(t *testing.T) {
	// No t.Parallel() - tests share global state and must run sequentially

	// Clear any existing children
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		activeChildren = nil
	}()

	// Register one process
	proc1 := &os.Process{Pid: 1000}
	Register(proc1)

	// Try to unregister a different process
	proc2 := &os.Process{Pid: 2000}
	Unregister(proc2)

	// Verify original process is still there
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 1 {
			t.Errorf("Expected 1 child process, got %d", len(activeChildren))
		}
		if activeChildren[0] != proc1 {
			t.Error("Original process should still be registered")
		}
	}()
}

//nolint:paralleltest // Tests share global state and must run sequentially
func TestKillAll(t *testing.T) {
	// No t.Parallel() - tests share global state and must run sequentially

	// Clear any existing children
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		activeChildren = nil
	}()

	// Register some dummy processes
	// Note: We use invalid PIDs so they can't actually be killed
	proc1 := &os.Process{Pid: 999999}
	proc2 := &os.Process{Pid: 999998}

	Register(proc1)
	Register(proc2)

	// Verify processes were registered
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 2 {
			t.Errorf("Expected 2 child processes, got %d", len(activeChildren))
		}
	}()

	// Call KillAll (will fail to kill since PIDs don't exist, but should clear list)
	KillAll()

	// Verify all processes were cleared
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 0 {
			t.Errorf("Expected 0 child processes after KillAll, got %d", len(activeChildren))
		}
	}()
}

//nolint:paralleltest // Tests share global state and must run sequentially
func TestKillAllWithNilProcess(t *testing.T) {
	// No t.Parallel() - tests share global state and must run sequentially

	// Clear any existing children
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		activeChildren = nil
	}()

	// Register a nil process (edge case)
	Register(nil)

	// Verify it was registered
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 1 {
			t.Errorf("Expected 1 child process, got %d", len(activeChildren))
		}
	}()

	// Call KillAll - should handle nil gracefully
	KillAll()

	// Verify list was cleared
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		if len(activeChildren) != 0 {
			t.Errorf("Expected 0 child processes after KillAll, got %d", len(activeChildren))
		}
	}()
}

//nolint:paralleltest // Tests share global state and must run sequentially
func TestConcurrentAccess(t *testing.T) {
	// No t.Parallel() - tests share global state and must run sequentially

	// Clear any existing children
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		activeChildren = nil
	}()

	// Test concurrent access to ensure proper mutex usage
	var waitGroup sync.WaitGroup
	waitGroup.Add(2)

	// Goroutine 1: Register processes
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 100; i++ {
			proc := &os.Process{Pid: i + 10000}
			Register(proc)
		}
	}()

	// Goroutine 2: Unregister processes
	go func() {
		defer waitGroup.Done()
		for i := 0; i < 50; i++ {
			proc := &os.Process{Pid: i + 10000}
			Unregister(proc)
		}
	}()

	// Wait for both goroutines
	waitGroup.Wait()

	// Just verify we didn't panic - the exact count doesn't matter
	func() {
		childrenMutex.Lock()
		defer childrenMutex.Unlock()
		count := len(activeChildren)

		if count < 0 || count > 100 {
			t.Errorf("Unexpected child count after concurrent access: %d", count)
		}
	}()
}
