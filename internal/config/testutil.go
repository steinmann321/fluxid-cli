package config

// Test helper functions for config package tests.

func strPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func equalHomeConfig(a, b *HomeConfig) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return equalStrPtr(a.Agent, b.Agent) &&
		equalIntPtr(a.Iterations, b.Iterations) &&
		equalIntPtr(a.ImplementRetries, b.ImplementRetries) &&
		equalBoolPtr(a.CommitEnabled, b.CommitEnabled)
}

func equalProjectConfig(a, b *ProjectConfig) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	return equalStrPtr(a.Agent, b.Agent) &&
		equalIntPtr(a.Iterations, b.Iterations) &&
		equalIntPtr(a.ImplementRetries, b.ImplementRetries) &&
		equalBoolPtr(a.CommitEnabled, b.CommitEnabled)
}

func equalStrPtr(a, b *string) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalIntPtr(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func equalBoolPtr(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
