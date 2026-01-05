//nolint:exhaustruct // Test file with partial config structures
package config

import (
	"strings"
	"testing"
)

//nolint:cyclop // Test function with multiple conditional paths
func TestValidateConfigCommitRetries(t *testing.T) {
	t.Parallel()

	// Test HomeConfig
	t.Run("HomeConfig", func(t *testing.T) {
		t.Parallel()

		err := validateHomeConfig(&HomeConfig{CommitRetries: intPtr(50)})
		if err != nil {
			t.Errorf("Unexpected error for valid commit retries: %v", err)
		}

		err = validateHomeConfig(&HomeConfig{CommitRetries: intPtr(0)})
		if err == nil || !strings.Contains(err.Error(), "commit_retries must be a positive integer") {
			t.Errorf("Expected error for zero commit retries, got: %v", err)
		}

		err = validateHomeConfig(&HomeConfig{CommitRetries: intPtr(-5)})
		if err == nil || !strings.Contains(err.Error(), "commit_retries must be a positive integer") {
			t.Errorf("Expected error for negative commit retries, got: %v", err)
		}
	})

	// Test ProjectConfig
	t.Run("ProjectConfig", func(t *testing.T) {
		t.Parallel()

		err := validateProjectConfig(&ProjectConfig{CommitRetries: intPtr(100)})
		if err != nil {
			t.Errorf("Unexpected error for valid commit retries: %v", err)
		}

		err = validateProjectConfig(&ProjectConfig{CommitRetries: intPtr(0)})
		if err == nil || !strings.Contains(err.Error(), "commit_retries must be a positive integer") {
			t.Errorf("Expected error for zero commit retries, got: %v", err)
		}

		err = validateProjectConfig(&ProjectConfig{CommitRetries: intPtr(-1)})
		if err == nil || !strings.Contains(err.Error(), "commit_retries must be a positive integer") {
			t.Errorf("Expected error for negative commit retries, got: %v", err)
		}
	})
}
