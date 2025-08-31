// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the configuration structure for restart-compose
type Config struct {
	Stacks []string `yaml:"stacks"`
}

// Find searches for .restart-compose.yaml in the specified directory
func Find(dir string) (string, error) {
	var searchDir string
	var err error

	if dir == "" {
		// If no directory specified, use current working directory
		searchDir, err = os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get current working directory: %w", err)
		}
	} else {
		// Use the specified directory
		searchDir, err = filepath.Abs(dir)
		if err != nil {
			return "", fmt.Errorf("failed to resolve directory path: %w", err)
		}

		// Check if directory exists
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return "", fmt.Errorf("specified directory does not exist: %s", dir)
		}
	}

	configPath := filepath.Join(searchDir, ".restart-compose.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("configuration file .restart-compose.yaml not found in directory: %s", searchDir)
	}

	return configPath, nil
}

// Load reads and parses the .restart-compose.yaml configuration file
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}
