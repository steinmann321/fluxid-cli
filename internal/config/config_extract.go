package config

// configValues holds project and home config values for a specific field.
type configValues[T any] struct {
	project *T
	home    *T
}

func extractAgentValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[string] {
	var values configValues[string]
	if projectConfig != nil {
		values.project = projectConfig.Agent
	}
	if homeConfig != nil {
		values.home = homeConfig.Agent
	}
	return values
}

func extractImplementRetriesValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[int] {
	var values configValues[int]
	if projectConfig != nil {
		values.project = projectConfig.ImplementRetries
	}
	if homeConfig != nil {
		values.home = homeConfig.ImplementRetries
	}
	return values
}

func extractCommitRetriesValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[int] {
	var values configValues[int]
	if projectConfig != nil {
		values.project = projectConfig.CommitRetries
	}
	if homeConfig != nil {
		values.home = homeConfig.CommitRetries
	}
	return values
}

func extractIterationsValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[int] {
	var values configValues[int]
	if projectConfig != nil {
		values.project = projectConfig.Iterations
	}
	if homeConfig != nil {
		values.home = homeConfig.Iterations
	}
	return values
}

func extractAgentArgsValues(
	projectConfig *ProjectConfig,
	homeConfig *HomeConfig,
) configValues[[]string] {
	var values configValues[[]string]
	if projectConfig != nil && len(projectConfig.AgentArgs) > 0 {
		values.project = &projectConfig.AgentArgs
	}
	if homeConfig != nil && len(homeConfig.AgentArgs) > 0 {
		values.home = &homeConfig.AgentArgs
	}
	return values
}

// resolveAgentArgs resolves agent args with precedence: project > home > empty.
// No hardcoded defaults - users must specify agent_args in their config file.
func resolveAgentArgs(projectArgs *[]string, homeArgs *[]string) []string {
	if projectArgs != nil {
		return *projectArgs
	}
	if homeArgs != nil {
		return *homeArgs
	}
	// No config specified - return empty slice
	return []string{}
}
