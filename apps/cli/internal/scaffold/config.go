package scaffold

import (
	"fmt"
	"os"
)

// LoadProjectConfig loads the interlock config from the given directory and validates
// that this is a valid interlock project.
func LoadProjectConfig(projectDir string) (*InterlockConfig, error) {
	config, err := LoadInterlockConfig(projectDir)
	if err != nil {
		return nil, fmt.Errorf("this doesn't appear to be an interlock project: %w", err)
	}

	return config, nil
}

// GetCurrentProjectConfig loads the interlock config from the current working directory.
func GetCurrentProjectConfig() (*InterlockConfig, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get current directory: %w", err)
	}

	return LoadProjectConfig(cwd)
}
