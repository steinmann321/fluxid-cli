//nolint:paralleltest // CLI argument parsing tests
package command

import (
	"os"
	"testing"
)

func TestParseArgs_ConfigFlag(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--config=custom.yaml"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIConfigPath == nil || *args.CLIConfigPath != "custom.yaml" {
		t.Errorf("Expected config path=custom.yaml, got %v", args.CLIConfigPath)
	}
}

func TestParseArgs_MultipleConfigFlags_Error(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--config=first.yaml", "--config=second.yaml"}
	_, err := ParseArgs()
	if err == nil {
		t.Error("Expected error for multiple --config flags")
	}

	if err.Error() != "multiple --config flags not allowed" {
		t.Errorf("Expected 'multiple --config flags not allowed' error, got: %v", err)
	}
}

func TestParseArgs_ImplementCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--implement-command=scripts/implement.sh"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIImplementCommand == nil || *args.CLIImplementCommand != "scripts/implement.sh" {
		t.Errorf("Expected implement command=scripts/implement.sh, got %v", args.CLIImplementCommand)
	}
}

func TestParseArgs_ReviewCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--review-command=scripts/review.sh"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIReviewCommand == nil || *args.CLIReviewCommand != "scripts/review.sh" {
		t.Errorf("Expected review command=scripts/review.sh, got %v", args.CLIReviewCommand)
	}
}

func TestParseArgs_CommitCommand(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--commit-command=scripts/commit.sh"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLICommitCommand == nil || *args.CLICommitCommand != "scripts/commit.sh" {
		t.Errorf("Expected commit command=scripts/commit.sh, got %v", args.CLICommitCommand)
	}
}

func TestParseArgs_AllNewFlags(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{
		"fluxid",
		"--config=myconfig.yaml",
		"--implement-command=impl.sh",
		"--review-command=review.sh",
		"--commit-command=commit.sh",
	}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIConfigPath == nil || *args.CLIConfigPath != "myconfig.yaml" {
		t.Errorf("Expected config path=myconfig.yaml, got %v", args.CLIConfigPath)
	}

	if args.CLIImplementCommand == nil || *args.CLIImplementCommand != "impl.sh" {
		t.Errorf("Expected implement command=impl.sh, got %v", args.CLIImplementCommand)
	}

	if args.CLIReviewCommand == nil || *args.CLIReviewCommand != "review.sh" {
		t.Errorf("Expected review command=review.sh, got %v", args.CLIReviewCommand)
	}

	if args.CLICommitCommand == nil || *args.CLICommitCommand != "commit.sh" {
		t.Errorf("Expected commit command=commit.sh, got %v", args.CLICommitCommand)
	}
}

func TestParseArgs_EmptyValueAfterEquals(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--config="}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Empty value should be allowed (will be validated later)
	if args.CLIConfigPath == nil || *args.CLIConfigPath != "" {
		t.Errorf("Expected empty config path, got %v", args.CLIConfigPath)
	}
}

func TestParseArgs_PathWithSpaces(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"fluxid", "--config=path with spaces/config.yaml"}
	args, err := ParseArgs()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if args.CLIConfigPath == nil || *args.CLIConfigPath != "path with spaces/config.yaml" {
		t.Errorf("Expected config path with spaces, got %v", args.CLIConfigPath)
	}
}
