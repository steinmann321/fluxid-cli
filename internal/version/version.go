// Package version provides version information for fluxid.
package version

import "fmt"

// Version information. These values are injected at build time via ldflags.
//
//nolint:gochecknoglobals // Version variables are set via ldflags at build time
var (
	// Version is the semantic version (e.g., "0.1.0", "1.2.3").
	Version = "dev"
	// GitCommit is the git commit hash.
	GitCommit = "unknown"
	// BuildDate is the build timestamp.
	BuildDate = "unknown"
)

// Get returns the version string with commit hash for easy tracking.
func Get() string {
	if GitCommit != "unknown" && len(GitCommit) > 7 {
		return fmt.Sprintf("%s+%s", Version, GitCommit[:7])
	}
	return Version
}

// Full returns detailed version information including git commit and build date.
func Full() string {
	return fmt.Sprintf("fluxid %s (commit: %s, built: %s)", Version, GitCommit, BuildDate)
}
