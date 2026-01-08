package version

import (
	"strings"
	"testing"
)

const (
	testVersion   = "1.2.3"
	testUnknown   = "unknown"
	testCommit    = "a1b2c3d4e5f6"
	testBuildDate = "2026-01-08T08:00:00Z"
)

//nolint:paralleltest // Cannot run in parallel - modifies global variables
func TestGet_WithShortCommit(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := GitCommit
	defer func() {
		Version = origVersion
		GitCommit = origCommit
	}()

	// Set test values
	Version = testVersion
	GitCommit = testCommit

	result := Get()

	expected := "1.2.3+a1b2c3d"
	if result != expected {
		t.Errorf("Get() = %q, want %q", result, expected)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global variables
func TestGet_WithUnknownCommit(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := GitCommit
	defer func() {
		Version = origVersion
		GitCommit = origCommit
	}()

	// Set test values
	Version = testVersion
	GitCommit = testUnknown

	result := Get()

	expected := testVersion
	if result != expected {
		t.Errorf("Get() with unknown commit = %q, want %q", result, expected)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global variables
func TestGet_WithShortCommitOnly(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := GitCommit
	defer func() {
		Version = origVersion
		GitCommit = origCommit
	}()

	// Set test values with commit shorter than 7 chars
	Version = testVersion
	GitCommit = "abc123"

	result := Get()

	expected := testVersion
	if result != expected {
		t.Errorf("Get() with short commit = %q, want %q", result, expected)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global variables
func TestFull(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := GitCommit
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		GitCommit = origCommit
		BuildDate = origBuildDate
	}()

	// Set test values
	Version = testVersion
	GitCommit = "a1b2c3d4"
	BuildDate = testBuildDate

	result := Full()

	if !strings.Contains(result, "fluxid") {
		t.Errorf("Full() should contain 'fluxid', got %q", result)
	}
	if !strings.Contains(result, Version) {
		t.Errorf("Full() should contain version %q, got %q", Version, result)
	}
	if !strings.Contains(result, GitCommit) {
		t.Errorf("Full() should contain commit %q, got %q", GitCommit, result)
	}
	if !strings.Contains(result, BuildDate) {
		t.Errorf("Full() should contain build date %q, got %q", BuildDate, result)
	}
}

//nolint:paralleltest // Cannot run in parallel - modifies global variables
func TestFull_WithDefaults(t *testing.T) {
	// Save original values
	origVersion := Version
	origCommit := GitCommit
	origBuildDate := BuildDate
	defer func() {
		Version = origVersion
		GitCommit = origCommit
		BuildDate = origBuildDate
	}()

	// Use default values
	Version = "dev"
	GitCommit = testUnknown
	BuildDate = testUnknown

	result := Full()

	expectedParts := []string{"fluxid", "dev", "unknown"}
	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Full() should contain %q, got %q", part, result)
		}
	}
}
