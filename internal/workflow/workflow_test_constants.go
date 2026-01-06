package workflow

// Test session ID constants for workflow tests.
// These constants prevent goconst violations for repeated UUIDs in tests.
const (
	testSessionReviewCommit  = "4c387bd3-ea45-499a-ad9a-cd8640493dbb"
	testSessionWorkflowError = "abde8fe8-69b5-4df9-9445-2e9169181b2d"
	testSessionPhaseRun      = "6eb0063a-d602-4976-a734-d5cd748331d8"
	testSessionRun           = "a039d23d-5a28-45a1-a0e8-f38542003100"
	testSessionWait          = "dbcffb56-2691-4d5b-aead-a645d23136ad"
)
