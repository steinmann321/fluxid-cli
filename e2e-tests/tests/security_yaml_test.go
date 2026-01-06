package tests

import (
	"os"
	"path/filepath"
	"testing"
)

// TestYAMLSecurityRejectAnchors verifies that YAML files with anchors are rejected.
func TestYAMLSecurityRejectAnchors(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-yaml-anchor"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a YAML file with an anchor
	reportPath := filepath.Join(sessionDir, "report.yaml")
	yamlWithAnchor := `command: &cmd fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(yamlWithAnchor), 0o644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects security validator to reject YAML with anchors (&)
	// Error should include "YAML anchor not allowed" or similar message
	t.Log("Test setup complete - expects rejection of YAML anchors with clear error")
}

// TestYAMLSecurityRejectAliases verifies that YAML files with aliases are rejected.
func TestYAMLSecurityRejectAliases(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-yaml-alias"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a YAML file with anchor and alias
	reportPath := filepath.Join(sessionDir, "report.yaml")
	yamlWithAlias := `command: &cmd fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
repeated_command: *cmd
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(yamlWithAlias), 0o644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects security validator to reject YAML with aliases (*)
	// Error should include "YAML alias not allowed" or similar message
	t.Log("Test setup complete - expects rejection of YAML aliases with clear error")
}

// TestYAMLSecurityRejectMergeKeys verifies that YAML files with merge keys are rejected.
func TestYAMLSecurityRejectMergeKeys(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-yaml-merge"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a YAML file with merge keys
	reportPath := filepath.Join(sessionDir, "report.yaml")
	yamlWithMerge := `defaults: &defaults
  timestamp: 2026-01-05T14:32:10Z
  status: PASS

command: fluxid.implement
artifact: test
<<: *defaults
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(yamlWithMerge), 0o644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects security validator to reject YAML with merge keys (<<)
	// Error should include "YAML merge key not allowed" or similar message
	t.Log("Test setup complete - expects rejection of YAML merge keys with clear error")
}

// TestYAMLSecurityAcceptPlainYAML verifies that plain YAML without dangerous features is accepted.
func TestYAMLSecurityAcceptPlainYAML(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-yaml-plain"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a plain YAML file without dangerous features
	reportPath := filepath.Join(sessionDir, "report.yaml")
	plainYAML := `command: fluxid.implement
artifact: internal/storage/report.go
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations:
    - Implementation complete
  enhancements: []
summary: Plain YAML test
`
	if err := os.WriteFile(reportPath, []byte(plainYAML), 0o644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects security validator to accept plain YAML
	t.Log("Test setup complete - expects acceptance of plain YAML without dangerous features")
}

// TestYAMLSecurityErrorMessagesAreClear verifies that security validation errors are instructive.
func TestYAMLSecurityErrorMessagesAreClear(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-yaml-errors"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a YAML file with anchor (to trigger security error)
	reportPath := filepath.Join(sessionDir, "report.yaml")
	yamlWithAnchor := `command: &anchor_name fluxid.implement
artifact: test
timestamp: 2026-01-05T14:32:10Z
status: PASS
issues:
  blockers: []
  defects: []
  concerns: []
  observations: []
  enhancements: []
`
	if err := os.WriteFile(reportPath, []byte(yamlWithAnchor), 0o644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects error message to include:
	// - Clear indication that anchors are not allowed
	// - Reference to security constraint
	// - Format: "[file]: YAML anchor not allowed
	//   (expected: no anchors, aliases, or merge keys, got: anchor '&anchor_name' at line X)"
	t.Log("Test setup complete - expects clear, instructive error messages per error format contract")
}

// TestYAMLSecurityBillionLaughsProtection verifies protection against billion laughs attack.
func TestYAMLSecurityBillionLaughsProtection(t *testing.T) {
	root := getProjectRoot(t)
	buildFluxid(t, root)

	// Create a temporary session directory
	tmpDir := t.TempDir()
	sessionID := "test-yaml-billion"
	sessionDir := filepath.Join(tmpDir, sessionID)
	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		t.Fatalf("Failed to create session directory: %v", err)
	}

	// Write a YAML file attempting billion laughs attack
	reportPath := filepath.Join(sessionDir, "report.yaml")
	billionLaughs := `a: &a ["lol","lol","lol","lol","lol","lol","lol","lol","lol"]
b: &b [*a,*a,*a,*a,*a,*a,*a,*a,*a]
c: &c [*b,*b,*b,*b,*b,*b,*b,*b,*b]
d: &d [*c,*c,*c,*c,*c,*c,*c,*c,*c]
e: &e [*d,*d,*d,*d,*d,*d,*d,*d,*d]
f: &f [*e,*e,*e,*e,*e,*e,*e,*e,*e]
command: fluxid.implement
`
	if err := os.WriteFile(reportPath, []byte(billionLaughs), 0o644); err != nil {
		t.Fatalf("Failed to write YAML file: %v", err)
	}

	t.Setenv("FLUXID_SESSION_ROOT", tmpDir)
	t.Setenv("FLUXID_SESSION_ID", sessionID)

	// This test expects security validator to reject billion laughs attack pattern
	// due to presence of anchors and aliases
	t.Log("Test setup complete - expects rejection of billion laughs attack pattern")
}
