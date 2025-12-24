// Package command implements the CLI definition layer for fluxid.
package command

import (
	"os"
)

// osEnv implements config.EnvGetter using os.Getenv.
type osEnv struct{}

// Getenv returns the value of an environment variable.
func (osEnv) Getenv(key string) string {
	return os.Getenv(key)
}
